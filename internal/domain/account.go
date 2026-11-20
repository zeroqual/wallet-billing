package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("account not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInvalidInput      = errors.New("invalid input")
	ErrAlreadyExists     = errors.New("account already exists")
)

type Account struct {
	ID        uuid.UUID
	OwnerID   string
	Balance   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
