package game

import (
	"context"
	"testing"
)

func newTestServiceWithGame() *Service {
	svc := NewService(nil)

	_, err := svc.CreateGame(context.Background(), "game-123", "user1", "user2", NewChessEngine())
	if err != nil {
		panic(err)
	}

	return svc
}

func activateGame(t *testing.T, svc *Service) {
	t.Helper()

	_, err := svc.JoinGame("game-123", "user1")
	if err != nil {
		t.Fatalf("user1 failed to join: %v", err)
	}

	_, err = svc.JoinGame("game-123", "user2")
	if err != nil {
		t.Fatalf("user2 failed to join: %v", err)
	}
}

func TestJoinGame(t *testing.T) {
	svc := newTestServiceWithGame()

	state1, err := svc.JoinGame("game-123", "user1")
	if err != nil {
		t.Fatalf("user1 join failed: %v", err)
	}

	if !state1.Player1Connected {
		t.Fatalf("expected player1 to be connected")
	}

	if state1.Status != string(StatusWaiting) {
		t.Fatalf("expected status waiting after first join, got %s", state1.Status)
	}

	state2, err := svc.JoinGame("game-123", "user2")
	if err != nil {
		t.Fatalf("user2 join failed: %v", err)
	}

	if !state2.Player2Connected {
		t.Fatalf("expected player2 to be connected")
	}

	if state2.Status != string(StatusActive) {
		t.Fatalf("expected status active after both joined, got %s", state2.Status)
	}
}

func TestJoinGame_Forbidden(t *testing.T) {
	svc := newTestServiceWithGame()

	_, err := svc.JoinGame("game-123", "intruder")
	if err == nil {
		t.Fatal("expected forbidden error, got nil")
	}

	if err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestMoveTurnOrder(t *testing.T) {
	svc := newTestServiceWithGame()
	activateGame(t, svc)

	_, _, err := svc.MakeMove(context.Background(), "game-123", "user2", "e7e5")
	if err == nil {
		t.Fatal("expected not your turn error, got nil")
	}

	if err != ErrNotYourTurn {
		t.Fatalf("expected ErrNotYourTurn, got %v", err)
	}

	state, result, err := svc.MakeMove(context.Background(), "game-123", "user1", "e2e4")
	if err != nil {
		t.Fatalf("expected valid move, got error: %v", err)
	}

	if result.Move != "e2e4" {
		t.Fatalf("expected move e2e4, got %s", result.Move)
	}

	if state.CurrentTurnUserID != "user2" {
		t.Fatalf("expected current turn to switch to user2, got %s", state.CurrentTurnUserID)
	}
}

func TestInvalidMove(t *testing.T) {
	svc := newTestServiceWithGame()
	activateGame(t, svc)

	_, _, err := svc.MakeMove(context.Background(), "game-123", "user1", "e2e5")
	if err == nil {
		t.Fatal("expected invalid move error, got nil")
	}

	if err != ErrInvalidMove {
		t.Fatalf("expected ErrInvalidMove, got %v", err)
	}
}

func TestCheckmate_FoolsMate(t *testing.T) {
	svc := newTestServiceWithGame()
	activateGame(t, svc)

	_, _, err := svc.MakeMove(context.Background(), "game-123", "user1", "f2f3")
	if err != nil {
		t.Fatalf("move f2f3 failed: %v", err)
	}

	_, _, err = svc.MakeMove(context.Background(), "game-123", "user2", "e7e5")
	if err != nil {
		t.Fatalf("move e7e5 failed: %v", err)
	}

	_, _, err = svc.MakeMove(context.Background(), "game-123", "user1", "g2g4")
	if err != nil {
		t.Fatalf("move g2g4 failed: %v", err)
	}

	state, result, err := svc.MakeMove(context.Background(), "game-123", "user2", "d8h4")
	if err != nil {
		t.Fatalf("move d8h4 failed: %v", err)
	}

	if !result.IsCheckmate {
		t.Fatal("expected checkmate to be detected")
	}

	if state.Status != string(StatusFinished) {
		t.Fatalf("expected game status finished, got %s", state.Status)
	}

	if state.WinnerID != "user2" {
		t.Fatalf("expected winner user2, got %s", state.WinnerID)
	}

	if state.FinishedReason != "checkmate" {
		t.Fatalf("expected finished reason checkmate, got %s", state.FinishedReason)
	}
}
