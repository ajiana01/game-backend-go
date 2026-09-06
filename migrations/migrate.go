package migrations

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed 001_init.sql
var initialSchema string

func Up(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, initialSchema)
	return err
}
