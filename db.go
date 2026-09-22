package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func DBPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(ctx, dsn)

	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)

	if err != nil {
		pool.Close()
		return nil, err
	}


	return pool, nil
}
