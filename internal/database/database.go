// Package database handles establishing and managing the database connection.
package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// New creates a new PostgreSQL connection pool.
func New(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// Use context.Background() for the initial connection setup.
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Ping the database to verify the connection is alive.
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close() // Clean up the pool if ping fails
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Successfully connected to the database!")
	return pool, nil
}
