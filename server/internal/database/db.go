package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	connectAttempts = 30
	connectDelay    = time.Second
)

func Open(ctx context.Context, databaseURL string) (*pgx.Conn, error) {
	var lastErr error

	for attempt := 1; attempt <= connectAttempts; attempt++ {
		conn, err := pgx.Connect(ctx, databaseURL)
		if err == nil {
			return conn, nil
		}

		lastErr = err
		log.Printf("database not ready yet (attempt %d/%d): %v", attempt, connectAttempts, err)
		if attempt < connectAttempts {
			time.Sleep(connectDelay)
		}
	}

	return nil, lastErr
}

func Connect(ctx context.Context, databaseURL string) *pgxpool.Pool {
	var lastErr error

	for attempt := 1; attempt <= connectAttempts; attempt++ {
		pool, err := pgxpool.New(ctx, databaseURL)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				log.Println("Connected to database")
				return pool
			}
			pool.Close()
		}

		lastErr = err
		log.Printf("database not ready yet (attempt %d/%d): %v", attempt, connectAttempts, err)
		if attempt < connectAttempts {
			time.Sleep(connectDelay)
		}
	}

	log.Fatalf("failed to connect to database: %v", lastErr)
	return nil
}
