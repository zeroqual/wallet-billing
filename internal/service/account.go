package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/zeroqual/wallet-billing/internal/domain"
)

type AccountRepository interface {
	Create(ctx context.Context, ownerID string) (domain.Account, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Account, error)
	Deposit(ctx context.Context, id uuid.UUID, amount int64) (domain.Account, error)
	Withdraw(ctx context.Context, id uuid.UUID, amount int64) (domain.Account, error)
}

type AccountService struct {
	repo AccountRepository
}

func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) Create(ctx context.Context, ownerID string) (domain.Account, error) {
	if ownerID == "" {
		return domain.Account{}, fmt.Errorf("%w: owner_id is empty", domain.ErrInvalidInput)
	}
	return s.repo.Create(ctx, ownerID)
}

func (s *AccountService) Get(ctx context.Context, id uuid.UUID) (domain.Account, error) {
	return s.repo.Get(ctx, id)
}

func (s *AccountService) Deposit(ctx context.Context, id uuid.UUID, amount int64) (domain.Account, error) {
	if amount <= 0 {
		return domain.Account{}, domain.ErrInvalidAmount
	}
	return s.repo.Deposit(ctx, id, amount)
}

func (s *AccountService) Withdraw(ctx context.Context, id uuid.UUID, amount int64) (domain.Account, error) {
	if amount <= 0 {
		return domain.Account{}, domain.ErrInvalidAmount
	}
	return s.repo.Withdraw(ctx, id, amount)
}
