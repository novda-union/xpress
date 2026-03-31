package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

	suffix := "." + direction + ".sql"
	entries, err := os.ReadDir("migrations")
	if err != nil {
		log.Fatalf("failed to read migrations directory: %v", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), suffix) {
			files = append(files, filepath.Join("migrations", e.Name()))
		}
	}
	sort.Strings(files)
	if direction == "down" {
		for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
			files[i], files[j] = files[j], files[i]
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
