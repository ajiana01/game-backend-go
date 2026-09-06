package leaderboard

import (
	"context"
	"sort"
	"testing"

	"github.com/ajiana01/portfolio-go/internal/domainerr"
	"github.com/ajiana01/portfolio-go/internal/player"
)

type memoryScores map[string]int64

func (m memoryScores) Submit(_ context.Context, id string, score int64) error {
	if score > m[id] {
		m[id] = score
	}
	return nil
}

func (m memoryScores) Top(_ context.Context, limit int64) ([]Score, error) {
	values := make([]Score, 0, len(m))
	for id, score := range m {
		values = append(values, Score{PlayerID: id, Score: score})
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Score > values[j].Score })
	if int64(len(values)) > limit {
		values = values[:limit]
	}
	return values, nil
}

type memoryPlayers map[string]player.Player

func (m memoryPlayers) Get(_ context.Context, id string) (player.Player, error) {
	value, ok := m[id]
	if !ok {
		return player.Player{}, domainerr.ErrNotFound
	}
	return value, nil
}

func TestSubmitKeepsHighestScoreAndTopAddsProfile(t *testing.T) {
	t.Parallel()
	scores := memoryScores{}
	players := memoryPlayers{"p1": {ID: "p1", Username: "hero"}}
	service := NewService(scores, players)

	if err := service.Submit(context.Background(), "p1", 500); err != nil {
		t.Fatal(err)
	}
	if err := service.Submit(context.Background(), "p1", 100); err != nil {
		t.Fatal(err)
	}
	entries, err := service.Top(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Score != 500 || entries[0].Username != "hero" {
		t.Fatalf("Top() = %+v", entries)
	}
}
