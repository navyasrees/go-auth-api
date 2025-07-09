package models

import (
	"auth-api/internal/config"
	"context"
	"fmt"
	"log"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	IsVerified   bool      `json:"is_verified"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

func CreateUser(email, passwordHash string) error {
	query := `
		INSERT INTO users (email, password_hash, is_verified, role, created_at)
		VALUES ($1, $2, false, 'user', NOW())
	`
	_, err := config.DB.Exec(context.Background(), query, email, passwordHash)
	return err
}

func StoreOTP(email, otp string) error {
    expires := time.Now().Add(15 * time.Minute)
    _, err := config.DB.Exec(context.Background(),
        `INSERT INTO otp_verifications (email, otp, expires_at) VALUES ($1, $2, $3)
         ON CONFLICT (email) DO UPDATE SET otp = $2, expires_at = $3`,
        email, otp, expires,
    )
    return err
}

func VerifyUserOTP(email, otp string) error {
    // First, check if OTP exists and is not expired
    var storedOTP string
    var expiresAt time.Time
    
    err := config.DB.QueryRow(context.Background(),
        `SELECT otp, expires_at FROM otp_verifications WHERE email = $1`,
        email,
    ).Scan(&storedOTP, &expiresAt)
    
    if err != nil {
        return err
    }
    
    // Check if OTP is expired
    if time.Now().After(expiresAt) {
        return fmt.Errorf("OTP has expired")
    }
    
    // Check if OTP matches
    if storedOTP != otp {
        return fmt.Errorf("invalid OTP")
    }
    
    // Update user verification status
    _, err = config.DB.Exec(context.Background(),
        `UPDATE users SET is_verified = true WHERE email = $1`,
        email,
    )
    if err != nil {
        return err
    }
    
    // Delete the used OTP
    _, err = config.DB.Exec(context.Background(),
        `DELETE FROM otp_verifications WHERE email = $1`,
        email,
    )
    
    return err
}

func GetUserByEmail(email string) (*User, error) {
	var user User
	err := config.DB.QueryRow(context.Background(),
		`SELECT id, email, password_hash, is_verified, role, created_at 
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsVerified, &user.Role, &user.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func GetUserByID(userID string) (*User, error) {
	var user User
	err := config.DB.QueryRow(context.Background(),
		`SELECT id, email, password_hash, is_verified, role, created_at 
		 FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsVerified, &user.Role, &user.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func StoreRefreshToken(userID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`
    log.Println("📬 query : ", query)
	_, err := config.DB.Exec(context.Background(), query, userID, tokenHash, expiresAt)
	return err
}

func GetRefreshTokenByHash(tokenHash string) (*RefreshToken, error) {
	var token RefreshToken
	err := config.DB.QueryRow(context.Background(),
		`SELECT id, user_id, token_hash, expires_at, created_at 
		 FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &token, nil
}

func DeleteRefreshToken(tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	_, err := config.DB.Exec(context.Background(), query, tokenHash)
	return err
}

func DeleteExpiredRefreshTokens() error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	_, err := config.DB.Exec(context.Background(), query)
	return err
}

type AccessToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func StoreAccessToken(userID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO access_tokens (user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := config.DB.Exec(context.Background(), query, userID, tokenHash, expiresAt)
	return err
}

func GetAccessTokenByHash(tokenHash string) (*AccessToken, error) {
	var token AccessToken
	err := config.DB.QueryRow(context.Background(),
		`SELECT id, user_id, token_hash, expires_at, created_at 
		 FROM access_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &token, nil
}

func DeleteAccessToken(tokenHash string) error {
	query := `DELETE FROM access_tokens WHERE token_hash = $1`
	_, err := config.DB.Exec(context.Background(), query, tokenHash)
	return err
}

func DeleteExpiredAccessTokens() error {
	query := `DELETE FROM access_tokens WHERE expires_at < NOW()`
	_, err := config.DB.Exec(context.Background(), query)
	return err
}

func DeleteAllAccessTokensForUser(userID string) error {
	query := `DELETE FROM access_tokens WHERE user_id = $1`
	_, err := config.DB.Exec(context.Background(), query, userID)
	return err
}

