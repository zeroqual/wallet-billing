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
	var account domain.Account
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Account{}, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	err = tx.QueryRow(ctx,
		`
			UPDATE accounts
			SET balance = balance+$1, updated_at = now()
			WHERE id = $2
			RETURNING id, owner_id, balance, created_at, updated_at
		`, amount, id,
	).Scan(&account.ID, &account.OwnerID, &account.Balance, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Account{}, fmt.Errorf("%w : %v", domain.ErrNotFound, err)
		}
		return domain.Account{}, fmt.Errorf("failed to update account balance: %w", err)
	}
	_, err = tx.Exec(ctx,
		`
			INSERT INTO transactions(account_id, type, amount, balance_after)
			VALUES($1, $2, $3, $4)
		`, account.ID, domain.TxDeposit, amount, account.Balance,
	)
	if err != nil {
		return domain.Account{}, fmt.Errorf("failed to insert into transactions: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Account{}, fmt.Errorf("failed to commit tx: %w", err)
	}
	return account, nil
}
func (r *AccountRepo) Withdraw(ctx context.Context, id uuid.UUID, amount int64) (domain.Account, error) {
	var account domain.Account
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Account{}, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	err = tx.QueryRow(ctx,
		`
			UPDATE accounts
			SET balance = balance-$1, updated_at = now()
			WHERE id = $2 AND balance-$1 >= 0
			RETURNING id, owner_id, balance, created_at, updated_at
		`, amount, id,
	).Scan(&account.ID, &account.OwnerID, &account.Balance, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			var exists bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)`, id).Scan(&exists); err != nil {
				return domain.Account{}, fmt.Errorf("failed to get account: %w", err)
			}
			if !exists {
				return domain.Account{}, domain.ErrNotFound
			}
			return domain.Account{}, domain.ErrInsufficientFunds
		}
		return domain.Account{}, fmt.Errorf("failed to update account balance: %w", err)
	}
	_, err = tx.Exec(ctx,
		`
			INSERT INTO transactions(account_id, type, amount, balance_after)
			VALUES($1, $2, $3, $4)
		`, account.ID, domain.TxWithdraw, amount, account.Balance,
	)
	if err != nil {
		return domain.Account{}, fmt.Errorf("failed to insert into transactions: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Account{}, fmt.Errorf("failed to commit tx: %w", err)
	}
	return account, nil
}
