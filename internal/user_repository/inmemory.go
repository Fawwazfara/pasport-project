package user_repository

import (
	"sync"
	"time"
)

type userRepository struct {
	mu    sync.RWMutex
	users map[string]User
}

func NewUserRepository() *userRepository {
	return &userRepository{users: make(map[string]User)}
}

func (r *userRepository) Create(username, fullName, passwordHash string) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[username]; ok {
		return User{}, ErrUsernameTaken
	}

	user := User{
		Username:     username,
		FullName:     fullName,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	r.users[username] = user
	return user, nil
}

func (r *userRepository) FindByUsername(username string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[username]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

func (r *userRepository) UpdatePassword(username, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[username]
	if !ok {
		return ErrUserNotFound
	}
	user.PasswordHash = passwordHash
	r.users[username] = user
	return nil
}
