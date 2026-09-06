package reward

import (
	"context"
	"strings"
	"time"

	"github.com/ajiana01/game-backend-go/internal/domainerr"
	"github.com/ajiana01/game-backend-go/internal/event"
	"github.com/ajiana01/game-backend-go/internal/inventory"
	"github.com/ajiana01/game-backend-go/internal/player"
)

type Type string

const (
	Login  Type = "login"
	Battle Type = "battle"
	Quest  Type = "quest"
)

type Reward struct {
	ID             string    `json:"id"`
	IdempotencyKey string    `json:"-"`
	PlayerID       string    `json:"player_id"`
	Type           Type      `json:"type"`
	Gold           int64     `json:"gold"`
	EXP            int64     `json:"exp"`
	Item           *Item     `json:"item,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Item struct {
	ItemID   string `json:"item_id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type GrantInput struct {
	Type Type `json:"type" enums:"login,battle,quest"`
}

type GrantResult struct {
	Reward  Reward        `json:"reward"`
	Player  player.Player `json:"player"`
	Created bool          `json:"-"`
}

type Repository interface {
	Grant(context.Context, string, string, Reward) (GrantResult, error)
}

type Inventory interface {
	AddReward(context.Context, string, string, inventory.AddInput) (inventory.Item, error)
}

type Service struct {
	repo      Repository
	inventory Inventory
	publisher event.Publisher
}

func NewService(repo Repository, inventory Inventory, publisher event.Publisher) *Service {
	return &Service{repo: repo, inventory: inventory, publisher: publisher}
}

func (s *Service) Grant(ctx context.Context, playerID, idempotencyKey string, rewardType Type) (GrantResult, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" || len(idempotencyKey) > 128 {
		return GrantResult{}, domainerr.ValidationError{Message: "Idempotency-Key header is required and must be at most 128 characters"}
	}
	template, ok := templates[rewardType]
	if !ok {
		return GrantResult{}, domainerr.ValidationError{Message: "type must be one of: login, battle, quest"}
	}
	result, err := s.repo.Grant(ctx, playerID, idempotencyKey, template)
	if err != nil {
		return GrantResult{}, err
	}
	if result.Reward.Item != nil {
		item := result.Reward.Item
		_, err = s.inventory.AddReward(ctx, playerID, result.Reward.ID, inventory.AddInput{
			ItemID: item.ItemID, Name: item.Name, Quantity: item.Quantity,
		})
		if err != nil {
			return GrantResult{}, err
		}
	}
	if result.Created {
		err = s.publisher.Publish(ctx, event.Event{
			ID: result.Reward.ID, Type: "reward.granted", OccurredAt: result.Reward.CreatedAt,
			Data: map[string]any{"player_id": playerID, "reward_type": rewardType, "gold": result.Reward.Gold, "exp": result.Reward.EXP},
		})
		if err != nil {
			return GrantResult{}, err
		}
	}
	return result, nil
}

var templates = map[Type]Reward{
	Login:  {Type: Login, Gold: 100, EXP: 25},
	Battle: {Type: Battle, Gold: 50, EXP: 100},
	Quest:  {Type: Quest, Gold: 200, EXP: 250, Item: &Item{ItemID: "health-potion", Name: "Health Potion", Quantity: 1}},
}
