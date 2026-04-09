package database

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

func Connect() *pgx.Conn {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	maxRetries := 30
	if v := os.Getenv("DB_MAX_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxRetries = n
		}
	}

	retryDelay := time.Second
	if v := os.Getenv("DB_RETRY_DELAY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			retryDelay = time.Duration(n) * time.Second
		}
	}

	var conn *pgx.Conn
	var err error
	for i := 0; i < maxRetries; i++ {
		conn, err = pgx.Connect(context.Background(), dbURL)
		if err == nil {
			return conn
		}
		log.Printf("database connect attempt %d/%d failed: %v; retrying in %s", i+1, maxRetries, err, retryDelay)
		time.Sleep(retryDelay)
	}

	log.Fatalf("failed to connect to database after %d attempts: %v", maxRetries, err)
	return nil
}
