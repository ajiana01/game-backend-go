package player

import (
	"context"
	"fmt"
	"github.com/ajiana01/portfolio-go/internal/domainerr"
	"regexp"
	"strings"
	"time"
)

var (
	usernameRE = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
)

type Player struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Level     int       `json:"level"`
	EXP       int64     `json:"exp"`
	Gold      int64     `json:"gold"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInput struct {
	Username string `json:"username"`
}

type UpdateInput struct {
	Username string `json:"username"`
}

type Repository interface {
	Create(context.Context, string) (Player, error)
	Get(context.Context, string) (Player, error)
	UpdateUsername(context.Context, string, string) (Player, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Register(ctx context.Context, input CreateInput) (Player, error) {
	username, err := validateUsername(input.Username)
	if err != nil {
		return Player{}, err
	}
	return s.repo.Create(ctx, username)
}

func (s *Service) Get(ctx context.Context, id string) (Player, error) {
	if strings.TrimSpace(id) == "" {
		return Player{}, domainerr.ValidationError{Message: "player id is required"}
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (Player, error) {
	username, err := validateUsername(input.Username)
	if err != nil {
		return Player{}, err
	}
	return s.repo.UpdateUsername(ctx, id, username)
}

func validateUsername(raw string) (string, error) {
	username := strings.TrimSpace(raw)
	if len(username) < 3 || len(username) > 24 {
		return "", domainerr.ValidationError{Message: "username must be between 3 and 24 characters"}
	}
	if !usernameRE.MatchString(username) {
		return "", domainerr.ValidationError{Message: "username may only contain letters, numbers, and underscores"}
	}
	return username, nil
}

func (p Player) String() string { return fmt.Sprintf("%s (%s)", p.Username, p.ID) }
