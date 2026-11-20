package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zeroqual/wallet-billing/internal/domain"
)

type AccountRepo struct {
	pool *pgxpool.Pool
}

func NewAccountRepo(pool *pgxpool.Pool) *AccountRepo { return &AccountRepo{pool: pool} }

func (r *AccountRepo) Create(ctx context.Context, ownerID string) (domain.Account, error) {
	var account domain.Account

	err := r.pool.QueryRow(ctx,
		`
			INSERT INTO accounts(owner_id)
			VALUES ($1)
			RETURNING id, owner_id, balance, created_at, updated_at
		`, ownerID,
	).Scan(&account.ID, &account.OwnerID, &account.Balance, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" {
			return domain.Account{}, fmt.Errorf("%w : %v", domain.ErrAlreadyExists, err)
		}
		return domain.Account{}, fmt.Errorf("create account: %w", err)
	}
	return account, nil
}
func (r *AccountRepo) Get(ctx context.Context, id uuid.UUID) (domain.Account, error) {
	var account domain.Account

	err := r.pool.QueryRow(ctx,
		`
			SELECT id, owner_id, balance, created_at, updated_at
			FROM accounts
			WHERE id = $1
			`, id,
	).Scan(&account.ID, &account.OwnerID, &account.Balance, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Account{}, fmt.Errorf("%w : %v", domain.ErrNotFound, err)
		}
		return domain.Account{}, fmt.Errorf("get account %s: %w", id, err)
	}
	return account, nil
}
func (r *AccountRepo) Deposit(ctx context.Context, id uuid.UUID, amount int64) (domain.Account, error) {
	return domain.Account{}, nil
}
func (r *AccountRepo) Withdraw(ctx context.Context, id uuid.UUID, amount int64) (domain.Account, error) {
	return domain.Account{}, nil
}
