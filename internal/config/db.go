package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB() {
    dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_NAME"),
    )

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v\n", err)
    }

    if err := pool.Ping(ctx); err != nil {
        log.Fatalf("Unable to ping database: %v\n", err)
    }

    DB = pool
    log.Println("📦 Connected to PostgreSQL")
}
