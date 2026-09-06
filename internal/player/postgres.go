package player

import (
	"context"
	"errors"

	"github.com/ajiana01/game-backend-go/internal/domainerr"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, username string) (Player, error) {
	const query = `INSERT INTO players (username) VALUES ($1)
		RETURNING id, username, level, exp, gold, created_at, updated_at`
	return scanPlayer(r.pool.QueryRow(ctx, query, username))
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (Player, error) {
	const query = `SELECT id, username, level, exp, gold, created_at, updated_at FROM players WHERE id = $1`
	return scanPlayer(r.pool.QueryRow(ctx, query, id))
}

func (r *PostgresRepository) Exists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM players WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) UpdateUsername(ctx context.Context, id, username string) (Player, error) {
	const query = `UPDATE players SET username = $2, updated_at = now() WHERE id = $1
		RETURNING id, username, level, exp, gold, created_at, updated_at`
	return scanPlayer(r.pool.QueryRow(ctx, query, id, username))
}

type rowScanner interface{ Scan(...any) error }

func scanPlayer(row rowScanner) (Player, error) {
	var value Player
	err := row.Scan(&value.ID, &value.Username, &value.Level, &value.EXP, &value.Gold, &value.CreatedAt, &value.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Player{}, domainerr.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Player{}, domainerr.ErrConflict
	}
	return value, err
}
