package repositories

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	*sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB() (*PostgresDB, error) {
	// Prefer DATABASE_URL when provided (common on Fly.io and other hosts)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		// Parse connection string to handle SSL mode correctly for Render
		// If using external URL (ends with .onrender.com), enforce sslmode=require
		// If using internal URL (avalon-db...), sslmode=disable is usually fine, but Render recommends require

		// Simple approach: trust the DATABASE_URL provided by platform but ensure sslmode is set if missing
		connStr := databaseURL
		if !strings.Contains(connStr, "sslmode=") {
			if strings.Contains(connStr, "?") {
				connStr += "&sslmode=require"
			} else {
				connStr += "?sslmode=require"
			}
		}

		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return nil, err
		}
		if err := db.Ping(); err != nil {
			return nil, err
		}
		return &PostgresDB{db}, nil
	}

	// Get connection parameters from environment variables
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "avalon")

	// Create connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresDB{db}, nil
}

// InitSchema initializes the database schema if it doesn't exist
func (db *PostgresDB) InitSchema() error {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create rooms table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS rooms (
			id UUID PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			host_id UUID NOT NULL REFERENCES users(id),
			status VARCHAR(20) NOT NULL,
			max_players INT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create game_sessions table with JSONB columns
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS game_sessions (
			id UUID PRIMARY KEY,
			room_id UUID NOT NULL REFERENCES rooms(id),
			game_type VARCHAR(50) NOT NULL,
			game_state JSONB NOT NULL,
			history JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create room_players join table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS room_users (
			room_id UUID REFERENCES rooms(id),
			user_id UUID REFERENCES users(id),
			joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (room_id, user_id)
		)
	`)

	return err
}

// Helper function to get environment variables with default fallback
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
