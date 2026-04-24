package pattern

import (
	"meme_chess/internal/analyzer/game"
	"meme_chess/internal/analyzer/movegen"
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
	"meme_chess/internal/analyzer/search"
)

type Detector struct {
	rules rules.RuleSet
}

func NewDetector(rs rules.RuleSet) *Detector {
	return &Detector{
		rules: rs,
	}
}

func (d *Detector) AnalyzeMove(gs *position.GameState, mv position.Move, ctx AnalysisContext) ([]Tag, int) {
	mover := gs.SideToMove
	movedPiece := gs.PieceAt(mv.From)
	beforeMaterial := materialBalance(gs, mover)

	if err := gs.ApplyMove(mv); err != nil {
		return nil, 0
	}
	defer gs.UndoMove()

	delta := ctx.Move.Score - ctx.BestScore
	afterMaterial := materialBalance(gs, mover)

	tags := make([]Tag, 0, 8)
	opponent := gs.SideToMove

	if mv.Kind == position.MoveCastleKingSide || mv.Kind == position.MoveCastleQueenSide {
		tags = append(tags, TagCastling)
	}

	if d.rules.IsCheck(gs, opponent) {
		tags = append(tags, TagCheck)
		if countCheckingPieces(gs, mover) >= 2 {
			tags = append(tags, TagDoubleCheck)
		}
	}

	if game.IsCheckmate(gs, d.rules) {
		tags = append(tags, TagCheckmate, TagForcedMate)
	}

	if ctx.Move.Score >= search.MateScore-8 {
		tags = append(tags, TagForcedMate)
	}

	if isFork(gs, mv.To, mover) {
		tags = append(tags, TagFork)
	}

	pinToKing, relativePin := pinTags(gs, mv.To, mover)
	if pinToKing {
		tags = append(tags, TagPinToKing)
	}
	if relativePin {
		tags = append(tags, TagRelativePin)
	}

	if createsAttack(gs, mv.To, mover) {
		tags = append(tags, TagAttack)
	}

	if isHangingAfterMove(gs, mv.To, movedPiece, mover) && delta <= -80 {
		tags = append(tags, TagHangingPiece)
	}

	if createsMateThreat(gs, mover) {
		tags = append(tags, TagMateThreat)
	}

	if isPerpetualCheck(gs, mover, ctx.Move.Score) {
		tags = append(tags, TagPerpetualCheck)
	}

	if afterMaterial-beforeMaterial >= 100 || createsWinningMaterial(gs, mover) {
		tags = append(tags, TagWinMaterial)
	}

	if ctx.BestScore >= 220 && (ctx.Move.Score >= 600 || ctx.Move.Score >= search.MateScore-8) {
		tags = append(tags, TagConversion)
	}

	if (mv.Kind == position.MoveCastleKingSide || mv.Kind == position.MoveCastleQueenSide) && d.rules.IsCheck(gs, opponent) {
		tags = append(tags, TagCastlingCheck)
	}

	if (mv.Kind == position.MoveCastleKingSide || mv.Kind == position.MoveCastleQueenSide) && (createsAttack(gs, mv.To, mover) || delta >= 120) {
		tags = append(tags, TagCastlingAttack)
	}

	tags = append(tags, openingTags(gs)...)

	return dedupe(tags), ctx.Delta
}

func openingTags(gs *position.GameState) []Tag {
	sequence := historyMoveKeys(gs)
	if len(sequence) < 4 {
		return nil
	}

	book := []struct {
		moves []string
		tags  []Tag
	}{
		{moves: []string{"e2e4", "e7e5", "g1f3", "b8c6"}, tags: []Tag{TagOpening, TagOpeningOpenGame}},
		{moves: []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "f8c5"}, tags: []Tag{TagOpening, TagOpeningItalian}},
		{moves: []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1b5"}, tags: []Tag{TagOpening, TagOpeningRuyLopez}},
		{moves: []string{"e2e4", "c7c5"}, tags: []Tag{TagOpening, TagOpeningSicilian}},
		{moves: []string{"e2e4", "e7e6"}, tags: []Tag{TagOpening, TagOpeningFrench}},
		{moves: []string{"e2e4", "c7c6"}, tags: []Tag{TagOpening, TagOpeningCaroKann}},
		{moves: []string{"d2d4", "d7d5", "c2c4"}, tags: []Tag{TagOpening, TagOpeningQueensGambit}},
		{moves: []string{"d2d4", "g8f6", "c2c4", "g7g6"}, tags: []Tag{TagOpening, TagOpeningKingsIndian}},
	}

	for _, entry := range book {
		if hasOpeningPrefix(sequence, entry.moves) {
			return entry.tags
		}
	}

	return nil
}

func hasOpeningPrefix(sequence, prefix []string) bool {
	if len(sequence) != len(prefix) {
		return false
	}
	for i := range prefix {
		if sequence[i] != prefix[i] {
			return false
		}
	}
	return true
}

func historyMoveKeys(gs *position.GameState) []string {
	out := make([]string, 0, len(gs.History))
	for _, item := range gs.History {
		out = append(out, moveKey(item.Move))
	}
	return out
}

func moveKey(mv position.Move) string {
	if mv.Promotion != position.NoPieceType {
		return mv.From.String() + mv.To.String() + promotionSuffix(mv.Promotion)
	}
	return mv.From.String() + mv.To.String()
}

func promotionSuffix(pt position.PieceType) string {
	switch pt {
	case position.Queen:
		return "q"
	case position.Rook:
		return "r"
	case position.Bishop:
		return "b"
	case position.Knight:
		return "n"
	default:
		return ""
	}
}

func isFork(gs *position.GameState, sq position.Square, mover position.Color) bool {
	piece := gs.PieceAt(sq)
	if piece.IsZero() || piece.Color != mover {
		return false
	}

	valuableTargets := 0

	for i := 0; i < 64; i++ {
		targetSq := position.Square(i)
		target := gs.PieceAt(targetSq)
		if target.IsZero() || target.Color == mover {
			continue
		}

		if !rules.AttacksSquare(gs, sq, targetSq, piece) {
			continue
		}

		if target.Type == position.King || pieceValue(target.Type) >= 300 {
			valuableTargets++
		}
	}

	return valuableTargets >= 2
}

func pinTags(gs *position.GameState, from position.Square, mover position.Color) (bool, bool) {
	piece := gs.PieceAt(from)
	if piece.IsZero() || piece.Color != mover {
		return false, false
	}

	pinToKing := false
	relativePin := false

	for _, dir := range slidingDirections(piece.Type) {
		file := from.File() + dir[0]
		rank := from.Rank() + dir[1]
		seenEnemy := position.NoSquare

		for file >= 0 && file <= 7 && rank >= 0 && rank <= 7 {
			sq := position.MustSquare(file, rank)
			target := gs.PieceAt(sq)

			if target.IsZero() {
				file += dir[0]
				rank += dir[1]
				continue
			}

			if target.Color == mover {
				break
			}

			if seenEnemy == position.NoSquare {
				seenEnemy = sq
				file += dir[0]
				rank += dir[1]
				continue
			}

			pinned := gs.PieceAt(seenEnemy)
			if target.Type == position.King {
				pinToKing = true
			} else if pieceValue(target.Type) > pieceValue(pinned.Type) {
				relativePin = true
			}
			break
		}
	}

	return pinToKing, relativePin
}

func slidingDirections(pt position.PieceType) [][2]int {
	switch pt {
	case position.Bishop:
		return [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	case position.Rook:
		return [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	case position.Queen:
		return [][2]int{
			{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
			{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		}
	default:
		return nil
	}
}

func createsAttack(gs *position.GameState, from position.Square, mover position.Color) bool {
	piece := gs.PieceAt(from)
	if piece.IsZero() || piece.Color != mover {
		return false
	}

	for i := 0; i < 64; i++ {
		targetSq := position.Square(i)
		target := gs.PieceAt(targetSq)
		if target.IsZero() || target.Color == mover {
			continue
		}

		if !rules.AttacksSquare(gs, from, targetSq, piece) {
			continue
		}

		if target.Type == position.King || pieceValue(target.Type) >= pieceValue(piece.Type) || isWeakTarget(gs, targetSq, target.Color) {
			return true
		}
	}

	return false
}

func isHangingAfterMove(gs *position.GameState, to position.Square, moved position.Piece, mover position.Color) bool {
	if moved.IsZero() || moved.Type == position.King {
		return false
	}

	return isWeakTarget(gs, to, mover)
}

func createsMateThreat(gs *position.GameState, attacker position.Color) bool {
	probe := gs.Clone()
	probe.SideToMove = attacker

	for _, mv := range dummiedLegalMoves(probe, attacker) {
		next := probe.Clone()
		if err := next.ApplyMove(mv); err != nil {
			continue
		}
		if game.IsCheckmate(next, rules.NewClassicalRuleSet()) {
			return true
		}
	}

	return false
}

func dummiedLegalMoves(gs *position.GameState, side position.Color) []position.Move {
	if gs.SideToMove != side {
		gs = gs.Clone()
		gs.SideToMove = side
	}
	return movegen.NewGenerator(rules.NewClassicalRuleSet()).GenerateLegalMoves(gs)
}

func createsWinningMaterial(gs *position.GameState, mover position.Color) bool {
	for i := 0; i < 64; i++ {
		sq := position.Square(i)
		p := gs.PieceAt(sq)
		if p.IsZero() || p.Color == mover || p.Type == position.King {
			continue
		}

		if isWeakTarget(gs, sq, p.Color) {
			return true
		}
	}

	return false
}

func isWeakTarget(gs *position.GameState, sq position.Square, color position.Color) bool {
	piece := gs.PieceAt(sq)
	if piece.IsZero() || piece.Color != color || piece.Type == position.King {
		return false
	}

	enemyAttackers := attackers(gs, sq, color.Opponent())
	if len(enemyAttackers) == 0 {
		return false
	}

	defenders := attackers(gs, sq, color)
	if len(defenders) == 0 {
		return true
	}

	return lowestAttackerValue(gs, enemyAttackers) < pieceValue(piece.Type)
}

func isPerpetualCheck(gs *position.GameState, attacker position.Color, moverScore int) bool {
	if !rules.NewClassicalRuleSet().IsCheck(gs, gs.SideToMove) {
		return false
	}
	if abs(moverScore) > 180 || moverScore >= 1000000-8 {
		return false
	}

	replies := dummiedLegalMoves(gs, gs.SideToMove)
	if len(replies) == 0 {
		return false
	}

	for _, reply := range replies {
		next := gs.Clone()
		if err := next.ApplyMove(reply); err != nil {
			continue
		}

		checkingReplyExists := false
		for _, mv := range dummiedLegalMoves(next, attacker) {
			after := next.Clone()
			if err := after.ApplyMove(mv); err != nil {
				continue
			}
			if rules.NewClassicalRuleSet().IsCheck(after, after.SideToMove) {
				checkingReplyExists = true
				break
			}
		}

		if !checkingReplyExists {
			return false
		}
	}

	return true
}

func countCheckingPieces(gs *position.GameState, attacker position.Color) int {
	kingSq, ok := findKing(gs, attacker.Opponent())
	if !ok {
		return 0
	}

	count := 0
	for i := 0; i < 64; i++ {
		from := position.Square(i)
		p := gs.PieceAt(from)
		if p.IsZero() || p.Color != attacker {
			continue
		}
		if rules.AttacksSquare(gs, from, kingSq, p) {
			count++
		}
	}

	return count
}

func attackers(gs *position.GameState, target position.Square, color position.Color) []position.Square {
	out := make([]position.Square, 0, 4)
	for i := 0; i < 64; i++ {
		from := position.Square(i)
		p := gs.PieceAt(from)
		if p.IsZero() || p.Color != color {
			continue
		}
		if rules.AttacksSquare(gs, from, target, p) {
			out = append(out, from)
		}
	}
	return out
}

func lowestAttackerValue(gs *position.GameState, squares []position.Square) int {
	best := 1 << 30
	for _, sq := range squares {
		value := pieceValue(gs.PieceAt(sq).Type)
		if value < best {
			best = value
		}
	}
	if best == 1<<30 {
		return 0
	}
	return best
}

func materialBalance(gs *position.GameState, perspective position.Color) int {
	score := 0
	for i := 0; i < 64; i++ {
		p := gs.PieceAt(position.Square(i))
		if p.IsZero() {
			continue
		}
		value := pieceValue(p.Type)
		if p.Color == perspective {
			score += value
		} else {
			score -= value
		}
	}
	return score
}

func findKing(gs *position.GameState, color position.Color) (position.Square, bool) {
	for i := 0; i < 64; i++ {
		sq := position.Square(i)
		p := gs.PieceAt(sq)
		if !p.IsZero() && p.Color == color && p.Type == position.King {
			return sq, true
		}
	}
	return position.NoSquare, false
}

func pieceValue(pt position.PieceType) int {
	switch pt {
	case position.Pawn:
		return 100
	case position.Knight:
		return 320
	case position.Bishop:
		return 330
	case position.Rook:
		return 500
	case position.Queen:
		return 900
	default:
		return 0
	}
}

func dedupe(tags []Tag) []Tag {
	seen := make(map[Tag]struct{}, len(tags))
	out := make([]Tag, 0, len(tags))
	for _, tag := range tags {
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
