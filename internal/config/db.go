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
    
    // Initialize database tables
    if err := initTables(); err != nil {
        log.Fatalf("Unable to initialize database tables: %v\n", err)
    }
}

func initTables() error {
    ctx := context.Background()
    
    // Create users table
    usersTable := `
        CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            email VARCHAR(255) UNIQUE NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            is_verified BOOLEAN DEFAULT FALSE,
            role VARCHAR(50) DEFAULT 'user',
            created_at TIMESTAMP DEFAULT NOW()
        )
    `
    
    // Create otp_verifications table
    otpVerificationsTable := `
        CREATE TABLE IF NOT EXISTS otp_verifications (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            email VARCHAR(255) UNIQUE NOT NULL,
            otp VARCHAR(10) NOT NULL,
            expires_at TIMESTAMP NOT NULL,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `
    
    // Create password_reset_otps table
    passwordResetOTPsTable := `
        CREATE TABLE IF NOT EXISTS password_reset_otps (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            email VARCHAR(255) UNIQUE NOT NULL,
            otp VARCHAR(10) NOT NULL,
            expires_at TIMESTAMP NOT NULL,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `
    
    // Create refresh_tokens table
    refreshTokensTable := `
        CREATE TABLE IF NOT EXISTS refresh_tokens (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            token_hash VARCHAR(255) UNIQUE NOT NULL,
            expires_at TIMESTAMP NOT NULL,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `
    
    // Create access_tokens table
    accessTokensTable := `
        CREATE TABLE IF NOT EXISTS access_tokens (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            token_hash VARCHAR(255) UNIQUE NOT NULL,
            expires_at TIMESTAMP NOT NULL,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `
    
    tables := []string{
        usersTable,
        otpVerificationsTable,
        passwordResetOTPsTable,
        refreshTokensTable,
        accessTokensTable,
    }
    
    for _, table := range tables {
        if _, err := DB.Exec(ctx, table); err != nil {
            return fmt.Errorf("failed to create table: %v", err)
        }
    }
    
    log.Println("📦 Database tables initialized successfully")
    return nil
}
