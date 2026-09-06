package inventory

import (
	"context"
	"strings"
	"time"

	"github.com/ajiana01/portfolio-go/internal/domainerr"
)

type Item struct {
	ItemID    string    `json:"item_id" bson:"item_id"`
	Name      string    `json:"name" bson:"name"`
	Quantity  int       `json:"quantity" bson:"quantity"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type AddInput struct {
	ItemID   string `json:"item_id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type RemoveInput struct {
	Quantity int `json:"quantity"`
}

type ListResponse struct {
	Items []Item `json:"items"`
}

type Repository interface {
	List(context.Context, string) ([]Item, error)
	Add(context.Context, string, AddInput) (Item, error)
	AddReward(context.Context, string, string, AddInput) (Item, error)
	Remove(context.Context, string, string, int) error
}

type PlayerLookup interface {
	Exists(context.Context, string) (bool, error)
}

type Service struct {
	repo    Repository
	players PlayerLookup
}

func NewService(repo Repository, players PlayerLookup) *Service {
	return &Service{repo: repo, players: players}
}

func (s *Service) List(ctx context.Context, playerID string) ([]Item, error) {
	if err := s.ensurePlayer(ctx, playerID); err != nil {
		return nil, err
	}
	items, err := s.repo.List(ctx, playerID)
	if items == nil {
		items = []Item{}
	}
	return items, err
}

func (s *Service) Add(ctx context.Context, playerID string, input AddInput) (Item, error) {
	input.ItemID = strings.TrimSpace(input.ItemID)
	input.Name = strings.TrimSpace(input.Name)
	if input.ItemID == "" || input.Name == "" {
		return Item{}, domainerr.ValidationError{Message: "item_id and name are required"}
	}
	if input.Quantity < 1 || input.Quantity > 9999 {
		return Item{}, domainerr.ValidationError{Message: "quantity must be between 1 and 9999"}
	}
	if err := s.ensurePlayer(ctx, playerID); err != nil {
		return Item{}, err
	}
	return s.repo.Add(ctx, playerID, input)
}

func (s *Service) AddReward(ctx context.Context, playerID, rewardID string, input AddInput) (Item, error) {
	if strings.TrimSpace(rewardID) == "" {
		return Item{}, domainerr.ValidationError{Message: "reward id is required"}
	}
	input.ItemID = strings.TrimSpace(input.ItemID)
	input.Name = strings.TrimSpace(input.Name)
	if input.ItemID == "" || input.Name == "" || input.Quantity < 1 {
		return Item{}, domainerr.ValidationError{Message: "reward item is invalid"}
	}
	return s.repo.AddReward(ctx, playerID, rewardID, input)
}

func (s *Service) Remove(ctx context.Context, playerID, itemID string, quantity int) error {
	if strings.TrimSpace(itemID) == "" || quantity < 1 {
		return domainerr.ValidationError{Message: "item id is required and quantity must be positive"}
	}
	if err := s.ensurePlayer(ctx, playerID); err != nil {
		return err
	}
	return s.repo.Remove(ctx, playerID, itemID, quantity)
}

func (s *Service) ensurePlayer(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return domainerr.ValidationError{Message: "player id is required"}
	}
	exists, err := s.players.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return domainerr.ErrNotFound
	}
	return nil
}
