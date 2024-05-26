package pool

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ranefattesingh/pkg/postgresql"
)

type Pool struct {
	pool *pgxpool.Pool
}

func NewDatabaseConnectionPool(ctx context.Context, connectionString string) (*Pool, error) {
	pool, err := pgxpool.New(ctx, postgresql.EncodeConnectionString(connectionString))
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	return &Pool{pool: pool}, nil
}

func (p *Pool) CloseDatabaseConnectionPool() {
	p.pool.Close()
}

func (p *Pool) Connection() *pgxpool.Pool {
	return p.pool
}
