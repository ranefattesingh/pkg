package pool

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	pool *pgxpool.Pool
}

func NewDatabaseConnectionPool(ctx context.Context, connectionString string) (*Pool, error) {
	pool, err := pgxpool.New(ctx, encodeConnectionString(connectionString))
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

func encodeConnectionString(connectionString string) string {
	urlBeginIndex := strings.Index(connectionString, "//")
	right := strings.LastIndex(connectionString, "@")
	left := urlBeginIndex + strings.Index(connectionString[urlBeginIndex:], ":") + 1
	password := connectionString[left:right]
	encodedPassword := url.QueryEscape(password)

	return strings.Replace(connectionString, password, encodedPassword, 1)
}
