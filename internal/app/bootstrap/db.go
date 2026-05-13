package bootstrap

import (
	"context"
	"fmt"
	"goph-profile/internal/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDBPool(ctx context.Context, cfg config.App) (*pgxpool.Pool, error) {
	url := cfg.DB.Conn
	if url == "" {
		return nil, fmt.Errorf("expected correct uri, found empty string instead")
	}

	pgxCfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("unable to parse uri: %w", err)
	}

	pgxCfg.MaxConnIdleTime = 5 * time.Second
	pgxCfg.MaxConnLifetime = 10 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping db: %w", err)
	}

	return pool, nil
}
