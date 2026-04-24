package analysis

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"meme_chess/internal/analyzer/pattern"

	_ "modernc.org/sqlite"
)

type Cache struct {
	db *sql.DB
}

func NewSQLiteCache() *Cache {
	cache, err := newCache("file:analysis-cache?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	return cache
}

func NewPersistentCache(path string) (*Cache, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create sqlite directory: %w", err)
	}
	return newCache(path)
}

func newCache(dsn string) (*Cache, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite cache: %w", err)
	}
	db.SetMaxOpenConns(1)

	cache := &Cache{db: db}
	if err := cache.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return cache, nil
}

func (c *Cache) init() error {
	stmts := []string{
		`PRAGMA foreign_keys = ON;`,
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA busy_timeout = 5000;`,
		`CREATE TABLE IF NOT EXISTS positions (
            position_hash TEXT PRIMARY KEY,
            depth INTEGER NOT NULL,
            tree_depth INTEGER NOT NULL DEFAULT 0,
            frontier_depth INTEGER NOT NULL DEFAULT 0,
            best_score_cp INTEGER NOT NULL,
            ready INTEGER NOT NULL
        );`,
		`CREATE TABLE IF NOT EXISTS moves (
            position_hash TEXT NOT NULL,
            move_key TEXT NOT NULL,
            score_cp INTEGER NOT NULL,
            delta_cp INTEGER NOT NULL,
            quality TEXT NOT NULL,
            tags_json TEXT NOT NULL,
            depth INTEGER NOT NULL,
            next_position_hash TEXT NOT NULL DEFAULT '',
            ready INTEGER NOT NULL,
            PRIMARY KEY (position_hash, move_key),
            FOREIGN KEY (position_hash) REFERENCES positions(position_hash) ON DELETE CASCADE
        );`,
	}

	for _, stmt := range stmts {
		if _, err := c.db.Exec(stmt); err != nil {
			return fmt.Errorf("init sqlite cache: %w", err)
		}
	}

	for _, migration := range []string{
		`ALTER TABLE positions ADD COLUMN tree_depth INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE positions ADD COLUMN frontier_depth INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE moves ADD COLUMN next_position_hash TEXT NOT NULL DEFAULT '';`,
	} {
		if _, err := c.db.Exec(migration); err != nil && !isDuplicateColumnError(err) {
			return fmt.Errorf("migrate sqlite cache: %w", err)
		}
	}

	return nil
}

func (c *Cache) Close() error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Close()
}

func (c *Cache) GetPosition(hash string) (*PositionAnalysis, bool) {
	row := c.db.QueryRow(`
        SELECT depth, tree_depth, frontier_depth, best_score_cp, ready
        FROM positions
        WHERE position_hash = ?
    `, hash)

	var p PositionAnalysis
	var ready int
	p.PositionHash = hash
	if err := row.Scan(&p.Depth, &p.TreeDepth, &p.FrontierDepth, &p.BestScoreCP, &ready); err != nil {
		return nil, false
	}
	p.Ready = ready == 1
	p.Moves = make(map[string]*MoveAnalysis)

	rows, err := c.db.Query(`
        SELECT move_key, score_cp, delta_cp, quality, tags_json, depth, next_position_hash, ready
        FROM moves
        WHERE position_hash = ?
    `, hash)
	if err != nil {
		return &p, true
	}
	defer rows.Close()

	for rows.Next() {
		var moveKey string
		move, err := scanMove(rows, &moveKey)
		if err != nil {
			continue
		}
		p.Moves[moveKey] = move
	}

	return &p, true
}

func (c *Cache) PutPosition(p *PositionAnalysis) {
	if p == nil {
		return
	}

	_, _ = c.db.Exec(`
        INSERT INTO positions(position_hash, depth, tree_depth, frontier_depth, best_score_cp, ready)
        VALUES(?, ?, ?, ?, ?, ?)
        ON CONFLICT(position_hash) DO UPDATE SET
            depth = CASE WHEN excluded.depth > positions.depth THEN excluded.depth ELSE positions.depth END,
            tree_depth = CASE WHEN excluded.tree_depth > positions.tree_depth THEN excluded.tree_depth ELSE positions.tree_depth END,
            frontier_depth = CASE WHEN excluded.frontier_depth > positions.frontier_depth THEN excluded.frontier_depth ELSE positions.frontier_depth END,
            best_score_cp = excluded.best_score_cp,
            ready = excluded.ready
    `, p.PositionHash, p.Depth, p.TreeDepth, p.FrontierDepth, p.BestScoreCP, boolToInt(p.Ready))
}

func (c *Cache) SetTreeDepth(hash string, treeDepth int) {
	_, _ = c.db.Exec(`
        UPDATE positions
        SET tree_depth = CASE WHEN ? > tree_depth THEN ? ELSE tree_depth END
        WHERE position_hash = ?
    `, treeDepth, treeDepth, hash)
}

func (c *Cache) SetFrontierDepth(hash string, frontierDepth int) {
	_, _ = c.db.Exec(`
        UPDATE positions
        SET frontier_depth = CASE WHEN ? > frontier_depth THEN ? ELSE frontier_depth END
        WHERE position_hash = ?
    `, frontierDepth, frontierDepth, hash)
}

func (c *Cache) GetMove(hash, moveKey string) (*MoveAnalysis, bool) {
	row := c.db.QueryRow(`
        SELECT score_cp, delta_cp, quality, tags_json, depth, next_position_hash, ready
        FROM moves
        WHERE position_hash = ? AND move_key = ?
    `, hash, moveKey)

	move, err := scanMove(row, nil)
	if err != nil {
		return nil, false
	}
	return move, true
}

func (c *Cache) PutMove(hash string, depth int, bestScore int, moveKey string, move *MoveAnalysis) {
	if move == nil {
		return
	}

	c.PutPosition(&PositionAnalysis{
		PositionHash:  hash,
		Depth:         depth,
		TreeDepth:     0,
		FrontierDepth: 0,
		BestScoreCP:   bestScore,
		Ready:         true,
	})

	tagsJSON, err := json.Marshal(move.Tags)
	if err != nil {
		tagsJSON = []byte("[]")
	}

	_, _ = c.db.Exec(`
        INSERT INTO moves(position_hash, move_key, score_cp, delta_cp, quality, tags_json, depth, next_position_hash, ready)
        VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(position_hash, move_key) DO UPDATE SET
            score_cp = excluded.score_cp,
            delta_cp = excluded.delta_cp,
            quality = excluded.quality,
            tags_json = excluded.tags_json,
            depth = CASE WHEN excluded.depth > moves.depth THEN excluded.depth ELSE moves.depth END,
            next_position_hash = excluded.next_position_hash,
            ready = excluded.ready
    `, hash, moveKey, move.ScoreCP, move.DeltaCP, move.Quality, string(tagsJSON), move.Depth, move.NextPositionHash, boolToInt(move.Ready))
}

func (c *Cache) Stats() (positions int, moves int) {
	_ = c.db.QueryRow(`SELECT COUNT(*) FROM positions`).Scan(&positions)
	_ = c.db.QueryRow(`SELECT COUNT(*) FROM moves`).Scan(&moves)
	return positions, moves
}

func (c *Cache) Dump() map[string]any {
	out := make(map[string]any)

	rows, err := c.db.Query(`
        SELECT p.position_hash, p.depth, p.tree_depth, p.frontier_depth, p.best_score_cp, p.ready,
               m.move_key, m.score_cp, m.delta_cp, m.quality, m.tags_json, m.depth, m.next_position_hash, m.ready
        FROM positions p
        LEFT JOIN moves m ON m.position_hash = p.position_hash
        ORDER BY p.position_hash, m.move_key
    `)
	if err != nil {
		return out
	}
	defer rows.Close()

	for rows.Next() {
		var (
			hash          string
			posDepth      int
			treeDepth     int
			frontierDepth int
			bestScore     int
			posReady      int
			moveKey       sql.NullString
			moveScore     sql.NullInt64
			moveDelta     sql.NullInt64
			moveQuality   sql.NullString
			tagsJSON      sql.NullString
			moveDepth     sql.NullInt64
			nextHash      sql.NullString
			moveReady     sql.NullInt64
		)

		if err := rows.Scan(&hash, &posDepth, &treeDepth, &frontierDepth, &bestScore, &posReady, &moveKey, &moveScore, &moveDelta, &moveQuality, &tagsJSON, &moveDepth, &nextHash, &moveReady); err != nil {
			continue
		}

		entry, ok := out[hash]
		if !ok {
			entry = map[string]any{
				"depth":          posDepth,
				"tree_depth":     treeDepth,
				"frontier_depth": frontierDepth,
				"best_score":     bestScore,
				"moves":          map[string]any{},
				"ready":          posReady == 1,
			}
			out[hash] = entry
		}

		if !moveKey.Valid {
			continue
		}

		tags, err := decodeTags(tagsJSON.String)
		if err != nil {
			tags = nil
		}

		entry.(map[string]any)["moves"].(map[string]any)[moveKey.String] = map[string]any{
			"score_cp":           int(moveScore.Int64),
			"delta_cp":           int(moveDelta.Int64),
			"quality":            moveQuality.String,
			"tags":               tags,
			"depth":              int(moveDepth.Int64),
			"next_position_hash": nextHash.String,
			"ready":              moveReady.Int64 == 1,
		}
	}

	return out
}

func scanMove(scanner interface{ Scan(dest ...any) error }, moveKey *string) (*MoveAnalysis, error) {
	var (
		maybeMoveKey     sql.NullString
		scoreCP          int
		deltaCP          int
		quality          string
		tagsJSON         string
		depth            int
		nextPositionHash string
		ready            int
	)

	if moveKey != nil {
		if err := scanner.Scan(&maybeMoveKey, &scoreCP, &deltaCP, &quality, &tagsJSON, &depth, &nextPositionHash, &ready); err != nil {
			return nil, err
		}
		*moveKey = maybeMoveKey.String
	} else {
		if err := scanner.Scan(&scoreCP, &deltaCP, &quality, &tagsJSON, &depth, &nextPositionHash, &ready); err != nil {
			return nil, err
		}
	}

	tags, err := decodeTags(tagsJSON)
	if err != nil {
		return nil, err
	}

	return &MoveAnalysis{
		ScoreCP:          scoreCP,
		DeltaCP:          deltaCP,
		Quality:          quality,
		Tags:             tags,
		Depth:            depth,
		NextPositionHash: nextPositionHash,
		Ready:            ready == 1,
	}, nil
}

func decodeTags(raw string) ([]pattern.Tag, error) {
	if raw == "" {
		return nil, nil
	}

	var tags []pattern.Tag
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func isDuplicateColumnError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate column name")
}
