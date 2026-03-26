package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/xpressgo/server/internal/config"
)

func main() {
	cfg := config.Load()

	conn, err := pgx.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}

	var file string
	if direction == "down" {
		file = "migrations/000001_init.down.sql"
	} else {
		file = "migrations/000001_init.up.sql"
	}

	sql, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("failed to read migration file: %v", err)
	}

	_, err = conn.Exec(context.Background(), string(sql))
	if err != nil {
		log.Fatalf("failed to run migration: %v", err)
	}

	log.Printf("Migration %s completed successfully", direction)
}
