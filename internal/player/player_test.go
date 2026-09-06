package player

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ajiana01/portfolio-go/internal/domainerr"
)

type memoryRepository struct{ players map[string]Player }

func (r *memoryRepository) Create(_ context.Context, username string) (Player, error) {
	for _, existing := range r.players {
		if existing.Username == username {
			return Player{}, domainerr.ErrConflict
		}
	}
	value := Player{ID: "player-1", Username: username, Level: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	r.players[value.ID] = value
	return value, nil
}

func (r *memoryRepository) Get(_ context.Context, id string) (Player, error) {
	value, ok := r.players[id]
	if !ok {
		return Player{}, domainerr.ErrNotFound
	}
	return value, nil
}

func (r *memoryRepository) UpdateUsername(_ context.Context, id, username string) (Player, error) {
	value, ok := r.players[id]
	if !ok {
		return Player{}, domainerr.ErrNotFound
	}
	value.Username = username
	r.players[id] = value
	return value, nil
}

func TestRegisterNormalizesAndCreatesPlayer(t *testing.T) {
	t.Parallel()
	repo := &memoryRepository{players: map[string]Player{}}
	service := NewService(repo)

	got, err := service.Register(context.Background(), CreateInput{Username: "  game_dev  "})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if got.Username != "game_dev" || got.Level != 1 {
		t.Fatalf("Register() = %+v", got)
	}
}

func TestRegisterRejectsInvalidUsername(t *testing.T) {
	t.Parallel()
	service := NewService(&memoryRepository{players: map[string]Player{}})

	_, err := service.Register(context.Background(), CreateInput{Username: "no spaces"})
	var validation domainerr.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("Register() error = %v, want validation error", err)
	}
}
