package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/xpressgo/server/internal/config"
	"github.com/xpressgo/server/internal/database"
)

func main() {
	cfg := config.Load()

	conn, err := database.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("failed to create schema_migrations table: %v", err)
	}

	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}

	var files []string
	if direction == "down" {
		files = []string{
			"migrations/000002_branches_and_permissions.down.sql",
			"migrations/000001_init.down.sql",
		}
	} else {
		files = []string{
			"migrations/000001_init.up.sql",
			"migrations/000002_branches_and_permissions.up.sql",
		}
	}

	for _, file := range files {
		name := filepath.Base(file)

		var already bool
		err = conn.QueryRow(context.Background(),
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)", name,
		).Scan(&already)
		if err != nil {
			log.Fatalf("failed to check migration %s: %v", name, err)
		}
		if already {
			log.Printf("skipping already-applied migration: %s", name)
			continue
		}

		sql, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("failed to read migration file %s: %v", file, err)
		}

		_, err = conn.Exec(context.Background(), string(sql))
		if err != nil {
			log.Fatalf("failed to run migration file %s: %v", file, err)
		}

		_, err = conn.Exec(context.Background(),
			"INSERT INTO schema_migrations (name) VALUES ($1)", name,
		)
		if err != nil {
			log.Fatalf("failed to record migration %s: %v", name, err)
		}

		log.Printf("applied migration: %s", name)
	}

	log.Printf("Migration %s completed successfully", direction)
}
