package user_repository

import (
	"errors"
	"time"
)

var (
	ErrUsernameTaken = errors.New("username sudah terdaftar")
	ErrUserNotFound  = errors.New("pengguna tidak ditemukan")
)

type User struct {
	Username     string
	FullName     string
	PasswordHash string
	CreatedAt    time.Time
}

type UserRepository interface {
	Create(username, fullName, passwordHash string) (User, error)
	FindByUsername(username string) (User, error)
	UpdatePassword(username, passwordHash string) error
}
