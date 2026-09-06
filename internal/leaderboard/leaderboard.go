package leaderboard

import (
	"context"
	"strings"

	"github.com/ajiana01/game-backend-go/internal/domainerr"
	"github.com/ajiana01/game-backend-go/internal/player"
)

type Score struct {
	PlayerID string `json:"player_id"`
	Score    int64  `json:"score"`
}

type Entry struct {
	Rank     int64  `json:"rank"`
	PlayerID string `json:"player_id"`
	Username string `json:"username"`
	Score    int64  `json:"score"`
}

type ListResponse struct {
	Players []Entry `json:"players"`
}

type Repository interface {
	Submit(context.Context, string, int64) error
	Top(context.Context, int64) ([]Score, error)
}

type PlayerReader interface {
	Get(context.Context, string) (player.Player, error)
}

type Service struct {
	repo    Repository
	players PlayerReader
}

func NewService(repo Repository, players PlayerReader) *Service {
	return &Service{repo: repo, players: players}
}

func (s *Service) Submit(ctx context.Context, playerID string, score int64) error {
	if strings.TrimSpace(playerID) == "" {
		return domainerr.ValidationError{Message: "player_id is required"}
	}
	if score < 0 {
		return domainerr.ValidationError{Message: "score must be zero or greater"}
	}
	if _, err := s.players.Get(ctx, playerID); err != nil {
		return err
	}
	return s.repo.Submit(ctx, playerID, score)
}

func (s *Service) Top(ctx context.Context, limit int64) ([]Entry, error) {
	if limit < 1 || limit > 100 {
		return nil, domainerr.ValidationError{Message: "limit must be between 1 and 100"}
	}
	scores, err := s.repo.Top(ctx, limit)
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, 0, len(scores))
	for index, score := range scores {
		value, err := s.players.Get(ctx, score.PlayerID)
		if err != nil {
			return nil, err
		}
		entries = append(entries, Entry{
			Rank: int64(index + 1), PlayerID: score.PlayerID, Username: value.Username, Score: score.Score,
		})
	}
	return entries, nil
}
