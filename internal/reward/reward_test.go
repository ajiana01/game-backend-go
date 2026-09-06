package reward

import (
	"context"
	"testing"
	"time"

	"github.com/ajiana01/game-backend-go/internal/event"
	"github.com/ajiana01/game-backend-go/internal/inventory"
	"github.com/ajiana01/game-backend-go/internal/player"
)

type memoryRewards struct{ values map[string]GrantResult }

func (r *memoryRewards) Grant(_ context.Context, playerID, key string, value Reward) (GrantResult, error) {
	if existing, ok := r.values[key]; ok {
		existing.Created = false
		return existing, nil
	}
	value.ID = "reward-1"
	value.PlayerID = playerID
	value.CreatedAt = time.Now()
	result := GrantResult{
		Reward:  value,
		Player:  player.Player{ID: playerID, Username: "hero", Gold: value.Gold, EXP: value.EXP, Level: 1},
		Created: true,
	}
	r.values[key] = result
	return result, nil
}

type rewardInventory struct {
	quantity int
	applied  map[string]bool
}

func (i *rewardInventory) AddReward(_ context.Context, _, rewardID string, input inventory.AddInput) (inventory.Item, error) {
	if !i.applied[rewardID] {
		i.quantity += input.Quantity
		i.applied[rewardID] = true
	}
	return inventory.Item{ItemID: input.ItemID, Name: input.Name, Quantity: i.quantity}, nil
}

type recordingPublisher struct{ events []event.Event }

func (p *recordingPublisher) Publish(_ context.Context, value event.Event) error {
	p.events = append(p.events, value)
	return nil
}

func TestQuestRetryDoesNotDuplicateEffects(t *testing.T) {
	t.Parallel()
	repo := &memoryRewards{values: map[string]GrantResult{}}
	items := &rewardInventory{applied: map[string]bool{}}
	publisher := &recordingPublisher{}
	service := NewService(repo, items, publisher)

	first, err := service.Grant(context.Background(), "player-1", "request-1", Quest)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Grant(context.Background(), "player-1", "request-1", Quest)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Created || second.Created {
		t.Fatalf("created flags = %v, %v", first.Created, second.Created)
	}
	if items.quantity != 1 {
		t.Fatalf("item quantity = %d, want 1", items.quantity)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("published events = %d, want 1", len(publisher.events))
	}
}
