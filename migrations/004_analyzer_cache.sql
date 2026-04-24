CREATE TABLE IF NOT EXISTS analysis_positions (
    position_hash TEXT PRIMARY KEY,
    depth INTEGER NOT NULL,
    tree_depth INTEGER NOT NULL DEFAULT 0,
    frontier_depth INTEGER NOT NULL DEFAULT 0,
    best_score_cp INTEGER NOT NULL,
    ready BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS analysis_moves (
    position_hash TEXT NOT NULL REFERENCES analysis_positions(position_hash) ON DELETE CASCADE,
    move_key TEXT NOT NULL,
    score_cp INTEGER NOT NULL,
    delta_cp INTEGER NOT NULL,
    quality TEXT NOT NULL,
    tags_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    depth INTEGER NOT NULL,
    next_position_hash TEXT NOT NULL DEFAULT '',
    ready BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (position_hash, move_key)
);

CREATE INDEX IF NOT EXISTS idx_analysis_moves_next_position_hash ON analysis_moves(next_position_hash);
