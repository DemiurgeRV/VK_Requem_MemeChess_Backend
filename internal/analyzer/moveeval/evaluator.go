package moveeval

import (
	"meme_chess/internal/analyzer/position"
	"meme_chess/internal/analyzer/rules"
)

var pawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	6, 8, 12, 24, 24, 12, 8, 6,
	2, 4, 8, 20, 20, 8, 4, 2,
	4, -2, -8, 0, 0, -8, -2, 4,
	4, 8, 8, -20, -20, 8, 8, 4,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightTable = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 4, 4, 0, -20, -40,
	-30, 4, 10, 15, 15, 10, 4, -30,
	-30, 8, 15, 20, 20, 15, 8, -30,
	-30, 4, 15, 20, 20, 15, 4, -30,
	-30, 8, 10, 15, 15, 10, 8, -30,
	-40, -20, 0, 8, 8, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopTable = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 4, 0, 0, 0, 0, 4, -10,
	-10, 8, 8, 10, 10, 8, 8, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 4, 10, 10, 10, 10, 4, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 6, 0, 0, 0, 0, 6, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookTable = [64]int{
	0, 0, 4, 10, 10, 4, 0, 0,
	-4, 0, 0, 0, 0, 0, 0, -4,
	-4, 0, 0, 0, 0, 0, 0, -4,
	-4, 0, 0, 0, 0, 0, 0, -4,
	-4, 0, 0, 0, 0, 0, 0, -4,
	-4, 0, 0, 0, 0, 0, 0, -4,
	4, 10, 10, 10, 10, 10, 10, 4,
	0, 0, 4, 10, 10, 4, 0, 0,
}

var queenTable = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 4, 0, 0, 0, 0, -10,
	-10, 4, 8, 8, 8, 8, 0, -10,
	0, 0, 8, 8, 8, 8, 0, -5,
	-5, 0, 8, 8, 8, 8, 0, -5,
	-10, 0, 8, 8, 8, 8, 0, -10,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingMidgameTable = [64]int{
	-40, -40, -40, -50, -50, -40, -40, -40,
	-30, -30, -30, -40, -40, -30, -30, -30,
	-20, -20, -20, -30, -30, -20, -20, -20,
	-10, -10, -10, -20, -20, -10, -10, -10,
	0, 0, -10, -20, -20, -10, 0, 0,
	10, 10, 0, -10, -10, 0, 10, 10,
	25, 25, 10, 0, 0, 10, 25, 25,
	25, 35, 15, 0, 0, 15, 35, 25,
}

var kingEndgameTable = [64]int{
	-40, -30, -20, -10, -10, -20, -30, -40,
	-30, -10, 0, 0, 0, 0, -10, -30,
	-20, 0, 10, 15, 15, 10, 0, -20,
	-10, 0, 15, 20, 20, 15, 0, -10,
	-10, 0, 15, 20, 20, 15, 0, -10,
	-20, 0, 10, 15, 15, 10, 0, -20,
	-30, -10, 0, 0, 0, 0, -10, -30,
	-40, -30, -20, -10, -10, -20, -30, -40,
}

func Evaluate(gs *position.GameState) int {
	score := rawEvaluation(gs)
	if gs.SideToMove == position.Black {
		score = -score
	}
	return score
}

func rawEvaluation(gs *position.GameState) int {
	score := 0
	pawnFiles := [2][8]int{}
	bishops := [2]int{}
	nonPawnMaterial := [2]int{}
	kingSquares := [2]position.Square{position.NoSquare, position.NoSquare}

	for i := 0; i < 64; i++ {
		sq := position.Square(i)
		p := gs.PieceAt(sq)
		if p.IsZero() {
			continue
		}

		side := colorIndex(p.Color)
		value := pieceValue(p.Type)
		score += signed(p.Color, value)
		score += signed(p.Color, pieceSquareBonus(p, sq, isEndgame(gs)))

		if p.Type == position.Pawn {
			pawnFiles[side][sq.File()]++
		} else if p.Type != position.King {
			nonPawnMaterial[side] += value
		}

		if p.Type == position.Bishop {
			bishops[side]++
		}
		if p.Type == position.King {
			kingSquares[side] = sq
		}
	}

	score += bishopPairBonus(bishops)
	score += pawnStructureScore(gs, pawnFiles)
	score += rookFileScore(gs, pawnFiles)
	score += developmentScore(gs)
	score += openingTheoryScore(gs)
	score += kingSafetyScore(gs, kingSquares, nonPawnMaterial)
	score += activityAndThreatScore(gs)

	return score
}

func activityAndThreatScore(gs *position.GameState) int {
	score := 0

	for i := 0; i < 64; i++ {
		from := position.Square(i)
		piece := gs.PieceAt(from)
		if piece.IsZero() {
			continue
		}

		attacks := 0
		pressure := 0
		for j := 0; j < 64; j++ {
			to := position.Square(j)
			if from == to {
				continue
			}

			target := gs.PieceAt(to)
			if !rules.AttacksSquare(gs, from, to, piece) {
				continue
			}

			attacks++
			if !target.IsZero() && target.Color != piece.Color {
				targetValue := pieceValue(target.Type)
				pressure += targetValue / 12
				if isHanging(gs, to, target.Color) {
					pressure += targetValue / 4
				}
				if pieceValue(piece.Type) < targetValue {
					pressure += 12
				}
			}
		}

		activity := attacks * mobilityWeight(piece.Type)
		score += signed(piece.Color, activity+pressure)

		if isHanging(gs, from, piece.Color) {
			penalty := pieceValue(piece.Type) / 3
			if piece.Type == position.Queen {
				penalty += 80
			}
			score -= signed(piece.Color, penalty)
		}
	}

	return score
}

func bishopPairBonus(bishops [2]int) int {
	score := 0
	if bishops[colorIndex(position.White)] >= 2 {
		score += 35
	}
	if bishops[colorIndex(position.Black)] >= 2 {
		score -= 35
	}
	return score
}

func pawnStructureScore(gs *position.GameState, pawnFiles [2][8]int) int {
	score := 0

	for i := 0; i < 64; i++ {
		sq := position.Square(i)
		p := gs.PieceAt(sq)
		if p.Type != position.Pawn {
			continue
		}

		side := colorIndex(p.Color)
		file := sq.File()

		if pawnFiles[side][file] > 1 {
			score -= signed(p.Color, 12*(pawnFiles[side][file]-1))
		}

		isolated := true
		for _, adj := range []int{file - 1, file + 1} {
			if adj >= 0 && adj < 8 && pawnFiles[side][adj] > 0 {
				isolated = false
				break
			}
		}
		if isolated {
			score -= signed(p.Color, 14)
		}

		if isPassedPawn(gs, sq, p.Color) {
			advance := sq.Rank()
			if p.Color == position.Black {
				advance = 7 - sq.Rank()
			}
			score += signed(p.Color, 18+advance*10)
		}
	}

	return score
}

func rookFileScore(gs *position.GameState, pawnFiles [2][8]int) int {
	score := 0

	for i := 0; i < 64; i++ {
		sq := position.Square(i)
		p := gs.PieceAt(sq)
		if p.Type != position.Rook {
			continue
		}

		file := sq.File()
		ownPawns := pawnFiles[colorIndex(p.Color)][file]
		enemyPawns := pawnFiles[colorIndex(p.Color.Opponent())][file]

		switch {
		case ownPawns == 0 && enemyPawns == 0:
			score += signed(p.Color, 24)
		case ownPawns == 0:
			score += signed(p.Color, 12)
		}

		if (p.Color == position.White && sq.Rank() == 6) || (p.Color == position.Black && sq.Rank() == 1) {
			score += signed(p.Color, 18)
		}
	}

	return score
}

func developmentScore(gs *position.GameState) int {
	if gs.FullmoveNumber > 18 {
		return 0
	}

	score := 0
	openingWeight := 19 - min(gs.FullmoveNumber, 18)

	for _, item := range []struct {
		sq    position.Square
		piece position.Piece
		bonus int
	}{
		{position.MustSquare(1, 0), position.Piece{Type: position.Knight, Color: position.White}, -16},
		{position.MustSquare(6, 0), position.Piece{Type: position.Knight, Color: position.White}, -16},
		{position.MustSquare(2, 0), position.Piece{Type: position.Bishop, Color: position.White}, -10},
		{position.MustSquare(5, 0), position.Piece{Type: position.Bishop, Color: position.White}, -10},
		{position.MustSquare(1, 7), position.Piece{Type: position.Knight, Color: position.Black}, 16},
		{position.MustSquare(6, 7), position.Piece{Type: position.Knight, Color: position.Black}, 16},
		{position.MustSquare(2, 7), position.Piece{Type: position.Bishop, Color: position.Black}, 10},
		{position.MustSquare(5, 7), position.Piece{Type: position.Bishop, Color: position.Black}, 10},
	} {
		if gs.PieceAt(item.sq) == item.piece {
			score += item.bonus * openingWeight / 6
		}
	}

	if isCastled(gs, position.White) {
		score += 28
	}
	if isCastled(gs, position.Black) {
		score -= 28
	}

	return score
}

func openingTheoryScore(gs *position.GameState) int {
	if len(gs.History) > 12 {
		return 0
	}

	score := 0

	// Central pawn presence and space in the opening.
	for _, item := range []struct {
		sq    position.Square
		piece position.Piece
		bonus int
	}{
		{position.MustSquare(4, 3), position.Piece{Type: position.Pawn, Color: position.White}, 42},  // e4
		{position.MustSquare(3, 3), position.Piece{Type: position.Pawn, Color: position.White}, 40},  // d4
		{position.MustSquare(2, 3), position.Piece{Type: position.Pawn, Color: position.White}, 20},  // c4
		{position.MustSquare(4, 4), position.Piece{Type: position.Pawn, Color: position.Black}, -42}, // e5
		{position.MustSquare(3, 4), position.Piece{Type: position.Pawn, Color: position.Black}, -40}, // d5
		{position.MustSquare(2, 4), position.Piece{Type: position.Pawn, Color: position.Black}, -20}, // c5
	} {
		if gs.PieceAt(item.sq) == item.piece {
			score += item.bonus
		}
	}

	// Slight bonus for healthy knight development, but avoid overvaluing it versus center pawns.
	for _, item := range []struct {
		sq    position.Square
		piece position.Piece
		bonus int
	}{
		{position.MustSquare(5, 2), position.Piece{Type: position.Knight, Color: position.White}, 18},  // Nf3
		{position.MustSquare(2, 2), position.Piece{Type: position.Knight, Color: position.White}, 14},  // Nc3
		{position.MustSquare(5, 5), position.Piece{Type: position.Knight, Color: position.Black}, -18}, // ...Nf6
		{position.MustSquare(2, 5), position.Piece{Type: position.Knight, Color: position.Black}, -14}, // ...Nc6
	} {
		if gs.PieceAt(item.sq) == item.piece {
			score += item.bonus
		}
	}

	// Penalize clearly anti-theoretical knight placement in the opening.
	for _, item := range []struct {
		sq    position.Square
		piece position.Piece
		bonus int
	}{
		{position.MustSquare(0, 2), position.Piece{Type: position.Knight, Color: position.White}, -38}, // Na3
		{position.MustSquare(7, 2), position.Piece{Type: position.Knight, Color: position.White}, -38}, // Nh3
		{position.MustSquare(0, 5), position.Piece{Type: position.Knight, Color: position.Black}, 38},  // ...Na6
		{position.MustSquare(7, 5), position.Piece{Type: position.Knight, Color: position.Black}, 38},  // ...Nh6
	} {
		if gs.PieceAt(item.sq) == item.piece {
			score += item.bonus
		}
	}

	// Reward occupying / controlling central squares with pawns and pieces.
	score += centerControlScore(gs)

	return score
}

func centerControlScore(gs *position.GameState) int {
	score := 0
	central := []position.Square{
		position.MustSquare(3, 3), // d4
		position.MustSquare(4, 3), // e4
		position.MustSquare(3, 4), // d5
		position.MustSquare(4, 4), // e5
	}

	for _, sq := range central {
		p := gs.PieceAt(sq)
		if !p.IsZero() {
			switch p.Type {
			case position.Pawn:
				score += signed(p.Color, 14)
			case position.Knight, position.Bishop:
				score += signed(p.Color, 8)
			}
		}

		for _, color := range []position.Color{position.White, position.Black} {
			attackers := attackersToSquare(gs, sq, color)
			score += signed(color, len(attackers)*3)
		}
	}

	return score
}

func attackersToSquare(gs *position.GameState, target position.Square, color position.Color) []position.Square {
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

func kingSafetyScore(gs *position.GameState, kingSquares [2]position.Square, nonPawnMaterial [2]int) int {
	score := 0
	endgame := isEndgame(gs)

	for _, color := range []position.Color{position.White, position.Black} {
		kingSq := kingSquares[colorIndex(color)]
		if kingSq == position.NoSquare {
			continue
		}

		if endgame {
			score += signed(color, pieceSquareValue(position.King, kingSq, color, true))
			continue
		}

		score += signed(color, kingShieldScore(gs, kingSq, color))

		if nonPawnMaterial[colorIndex(color)] > 1800 && kingSq.File() >= 3 && kingSq.File() <= 4 {
			score -= signed(color, 22)
		}

		enemyPressure := 0
		for _, near := range kingRing(kingSq) {
			enemyPressure += len(attackers(gs, near, color.Opponent()))
		}
		score -= signed(color, enemyPressure*8)
	}

	return score
}

func pieceSquareBonus(p position.Piece, sq position.Square, endgame bool) int {
	return pieceSquareValue(p.Type, sq, p.Color, endgame)
}

func pieceSquareValue(pt position.PieceType, sq position.Square, color position.Color, endgame bool) int {
	index := tableIndex(sq, color)

	switch pt {
	case position.Pawn:
		return pawnTable[index]
	case position.Knight:
		return knightTable[index]
	case position.Bishop:
		return bishopTable[index]
	case position.Rook:
		return rookTable[index]
	case position.Queen:
		return queenTable[index]
	case position.King:
		if endgame {
			return kingEndgameTable[index]
		}
		return kingMidgameTable[index]
	default:
		return 0
	}
}

func isPassedPawn(gs *position.GameState, sq position.Square, color position.Color) bool {
	file := sq.File()
	rank := sq.Rank()

	for i := 0; i < 64; i++ {
		otherSq := position.Square(i)
		p := gs.PieceAt(otherSq)
		if p.Type != position.Pawn || p.Color == color {
			continue
		}

		if abs(otherSq.File()-file) > 1 {
			continue
		}

		if color == position.White && otherSq.Rank() > rank {
			return false
		}
		if color == position.Black && otherSq.Rank() < rank {
			return false
		}
	}

	return true
}

func isHanging(gs *position.GameState, sq position.Square, color position.Color) bool {
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

	lowestEnemy := lowestAttackerValue(gs, enemyAttackers)
	return lowestEnemy < pieceValue(piece.Type)
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

func kingShieldScore(gs *position.GameState, kingSq position.Square, color position.Color) int {
	score := 0
	dir := 1
	homeRank := 0
	if color == position.Black {
		dir = -1
		homeRank = 7
	}

	if kingSq.Rank() == homeRank {
		score -= 12
	}

	for df := -1; df <= 1; df++ {
		file := kingSq.File() + df
		rank := kingSq.Rank() + dir
		if file < 0 || file > 7 || rank < 0 || rank > 7 {
			continue
		}

		pawn := gs.PieceAt(position.MustSquare(file, rank))
		if pawn.Type == position.Pawn && pawn.Color == color {
			score += 10
		} else {
			score -= 10
		}
	}

	return score
}

func kingRing(sq position.Square) []position.Square {
	out := make([]position.Square, 0, 8)
	for df := -1; df <= 1; df++ {
		for dr := -1; dr <= 1; dr++ {
			if df == 0 && dr == 0 {
				continue
			}
			file := sq.File() + df
			rank := sq.Rank() + dr
			if file < 0 || file > 7 || rank < 0 || rank > 7 {
				continue
			}
			out = append(out, position.MustSquare(file, rank))
		}
	}
	return out
}

func isCastled(gs *position.GameState, color position.Color) bool {
	if color == position.White {
		return gs.PieceAt(position.MustSquare(6, 0)) == (position.Piece{Type: position.King, Color: color}) ||
			gs.PieceAt(position.MustSquare(2, 0)) == (position.Piece{Type: position.King, Color: color})
	}
	return gs.PieceAt(position.MustSquare(6, 7)) == (position.Piece{Type: position.King, Color: color}) ||
		gs.PieceAt(position.MustSquare(2, 7)) == (position.Piece{Type: position.King, Color: color})
}

func isEndgame(gs *position.GameState) bool {
	nonPawnMaterial := 0
	queens := 0

	for i := 0; i < 64; i++ {
		p := gs.PieceAt(position.Square(i))
		if p.IsZero() || p.Type == position.Pawn || p.Type == position.King {
			continue
		}
		nonPawnMaterial += pieceValue(p.Type)
		if p.Type == position.Queen {
			queens++
		}
	}

	return queens == 0 || nonPawnMaterial <= 2600
}

func mobilityWeight(pt position.PieceType) int {
	switch pt {
	case position.Pawn:
		return 1
	case position.Knight, position.Bishop:
		return 2
	case position.Rook:
		return 2
	case position.Queen:
		return 1
	default:
		return 0
	}
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
	case position.King:
		return 0
	default:
		return 0
	}
}

func tableIndex(sq position.Square, color position.Color) int {
	if color == position.White {
		return int(sq)
	}
	return (7-sq.Rank())*8 + sq.File()
}

func signed(color position.Color, value int) int {
	if color == position.White {
		return value
	}
	return -value
}

func colorIndex(color position.Color) int {
	if color == position.White {
		return 0
	}
	return 1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
