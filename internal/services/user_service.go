package services

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"avalon/internal/repositories"

	"golang.org/x/crypto/argon2"
)

// UserService handles user-related business logic
type UserService struct {
	repo repositories.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// RegisterUser registers a new user
func (s *UserService) RegisterUser(username, password string) (*repositories.User, error) {
	// Validate input
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}

	// Check if username already exists
	existingUser, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Hash the password
	hash, err := generatePasswordHash(password)
	if err != nil {
		return nil, err
	}

	// Create the user
	user, err := s.repo.Create(username, hash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser authenticates a user
func (s *UserService) AuthenticateUser(username, password string) (*repositories.User, error) {
	// Get user by username
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	// Verify password
	match, err := verifyPassword(user.PasswordHash, password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

// GetUser gets a user by ID
func (s *UserService) GetUser(id string) (*repositories.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// UpdateUser updates a user's information
func (s *UserService) UpdateUser(id, username, password string) (*repositories.User, error) {
	// Get existing user
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Update username if provided
	if username != "" && username != user.Username {
		// Check if username already exists
		existingUser, err := s.repo.GetByUsername(username)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, errors.New("username already exists")
		}
		user.Username = username
	}

	// Update password if provided
	if password != "" {
		hash, err := generatePasswordHash(password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hash
	}

	// Update the user
	err = s.repo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id string) error {
	return s.repo.Delete(id)
}

// Password hashing parameters
const (
	saltLength = 16
	keyLength  = 32
	iterations = 3
	memory     = 64 * 1024
	threads    = 4
)

// generatePasswordHash generates an Argon2id hash from a password
func generatePasswordHash(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Generate the hash
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, threads, keyLength)

	// Encode as base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format the hash
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, threads, b64Salt, b64Hash)

	return encodedHash, nil
}

// verifyPassword verifies a password against an Argon2id hash
func verifyPassword(encodedHash, password string) (bool, error) {
	// Parse the hash
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return false, errors.New("invalid hash format")
	}

	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, errors.New("incompatible argon2 version")
	}

	var memory, iterations, threads int
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &memory, &iterations, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return false, err
	}

	// Generate a hash from the provided password with the same parameters
	hash := argon2.IDKey([]byte(password), salt, uint32(iterations), uint32(memory), uint8(threads), uint32(len(decodedHash)))

	// Compare the hashes
	return subtle.ConstantTimeCompare(hash, decodedHash) == 1, nil
}
