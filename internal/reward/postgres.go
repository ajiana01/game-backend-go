package reward

import (
	"context"
	"errors"

	"github.com/ajiana01/portfolio-go/internal/domainerr"
	"github.com/ajiana01/portfolio-go/internal/player"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Grant(ctx context.Context, playerID, key string, value Reward) (GrantResult, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return GrantResult{}, err
	}
	defer tx.Rollback(ctx)

	var itemID, itemName *string
	var itemQuantity int
	if value.Item != nil {
		itemID, itemName, itemQuantity = &value.Item.ItemID, &value.Item.Name, value.Item.Quantity
	}
	const insert = `INSERT INTO rewards (idempotency_key, player_id, type, gold, exp, item_id, item_name, item_quantity)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (idempotency_key) DO NOTHING
		RETURNING id, idempotency_key, player_id, type, gold, exp, item_id, item_name, item_quantity, created_at`
	created := true
	granted, err := scanReward(tx.QueryRow(ctx, insert, key, playerID, value.Type, value.Gold, value.EXP, itemID, itemName, itemQuantity))
	if errors.Is(err, pgx.ErrNoRows) {
		created = false
		const existing = `SELECT id, idempotency_key, player_id, type, gold, exp, item_id, item_name, item_quantity, created_at
			FROM rewards WHERE idempotency_key = $1`
		granted, err = scanReward(tx.QueryRow(ctx, existing, key))
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return GrantResult{}, domainerr.ErrNotFound
		}
		return GrantResult{}, err
	}
	if granted.PlayerID != playerID {
		return GrantResult{}, domainerr.ErrConflict
	}
	if granted.Type != value.Type {
		return GrantResult{}, domainerr.ErrConflict
	}

	var updated player.Player
	if created {
		const update = `UPDATE players SET gold = gold + $2, exp = exp + $3,
			level = 1 + ((exp + $3) / 1000), updated_at = now() WHERE id = $1
			RETURNING id, username, level, exp, gold, created_at, updated_at`
		err = tx.QueryRow(ctx, update, playerID, granted.Gold, granted.EXP).Scan(
			&updated.ID, &updated.Username, &updated.Level, &updated.EXP, &updated.Gold, &updated.CreatedAt, &updated.UpdatedAt,
		)
	} else {
		const get = `SELECT id, username, level, exp, gold, created_at, updated_at FROM players WHERE id = $1`
		err = tx.QueryRow(ctx, get, playerID).Scan(
			&updated.ID, &updated.Username, &updated.Level, &updated.EXP, &updated.Gold, &updated.CreatedAt, &updated.UpdatedAt,
		)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return GrantResult{}, domainerr.ErrNotFound
	}
	if err != nil {
		return GrantResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return GrantResult{}, err
	}
	return GrantResult{Reward: granted, Player: updated, Created: created}, nil
}

func scanReward(row interface{ Scan(...any) error }) (Reward, error) {
	var value Reward
	var itemID, itemName *string
	var itemQuantity int
	err := row.Scan(&value.ID, &value.IdempotencyKey, &value.PlayerID, &value.Type, &value.Gold, &value.EXP, &itemID, &itemName, &itemQuantity, &value.CreatedAt)
	if itemID != nil && itemName != nil {
		value.Item = &Item{ItemID: *itemID, Name: *itemName, Quantity: itemQuantity}
	}
	return value, err
}
