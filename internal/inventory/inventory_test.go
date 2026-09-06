package inventory

import (
	"context"
	"testing"

	"github.com/ajiana01/portfolio-go/internal/domainerr"
)

type memoryInventory struct {
	items   map[string]Item
	rewards map[string]struct{}
}

func (r *memoryInventory) List(context.Context, string) ([]Item, error) {
	values := make([]Item, 0, len(r.items))
	for _, item := range r.items {
		values = append(values, item)
	}
	return values, nil
}

func (r *memoryInventory) Add(_ context.Context, _ string, input AddInput) (Item, error) {
	value := r.items[input.ItemID]
	value.ItemID, value.Name, value.Quantity = input.ItemID, input.Name, value.Quantity+input.Quantity
	r.items[input.ItemID] = value
	return value, nil
}

func (r *memoryInventory) AddReward(ctx context.Context, playerID, rewardID string, input AddInput) (Item, error) {
	if _, exists := r.rewards[rewardID]; exists {
		return r.items[input.ItemID], nil
	}
	r.rewards[rewardID] = struct{}{}
	return r.Add(ctx, playerID, input)
}

func (r *memoryInventory) Remove(_ context.Context, _, itemID string, quantity int) error {
	value, ok := r.items[itemID]
	if !ok || value.Quantity < quantity {
		return domainerr.ErrNotFound
	}
	value.Quantity -= quantity
	if value.Quantity == 0 {
		delete(r.items, itemID)
	} else {
		r.items[itemID] = value
	}
	return nil
}

type playerExists bool

func (p playerExists) Exists(context.Context, string) (bool, error) { return bool(p), nil }

func TestRewardItemIsIdempotent(t *testing.T) {
	t.Parallel()
	repo := &memoryInventory{items: map[string]Item{}, rewards: map[string]struct{}{}}
	service := NewService(repo, playerExists(true))
	input := AddInput{ItemID: "potion", Name: "Potion", Quantity: 1}

	if _, err := service.AddReward(context.Background(), "player-1", "reward-1", input); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddReward(context.Background(), "player-1", "reward-1", input); err != nil {
		t.Fatal(err)
	}
	if got := repo.items["potion"].Quantity; got != 1 {
		t.Fatalf("quantity = %d, want 1", got)
	}
}

func TestAddRejectsMissingPlayer(t *testing.T) {
	t.Parallel()
	service := NewService(&memoryInventory{items: map[string]Item{}}, playerExists(false))
	_, err := service.Add(context.Background(), "missing", AddInput{ItemID: "potion", Name: "Potion", Quantity: 1})
	if err != domainerr.ErrNotFound {
		t.Fatalf("Add() error = %v, want not found", err)
	}
}
