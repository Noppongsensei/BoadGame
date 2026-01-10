package repositories

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(username, passwordHash string) (*User, error)
	GetByID(id string) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id string) error
}

// PostgresUserRepository implements UserRepository for PostgreSQL
type PostgresUserRepository struct {
	db *PostgresDB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *PostgresDB) UserRepository {
	return &PostgresUserRepository{db: db}
}

// Create adds a new user to the database
func (r *PostgresUserRepository) Create(username, passwordHash string) (*User, error) {
	user := &User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := r.db.Exec(
		`INSERT INTO users (id, username, password_hash, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(id string) (*User, error) {
	var user User
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *PostgresUserRepository) GetByUsername(username string) (*User, error) {
	var user User
	err := r.db.QueryRow(
		`SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update updates user information
func (r *PostgresUserRepository) Update(user *User) error {
	user.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		`UPDATE users SET username = $1, password_hash = $2, updated_at = $3 WHERE id = $4`,
		user.Username, user.PasswordHash, user.UpdatedAt, user.ID,
	)
	return err
}

// Delete removes a user from the database
func (r *PostgresUserRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	return err
}
