package postgres

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/zeroqual/wallet-billing/internal/domain"
)

func newTestRepo(t *testing.T) (*AccountRepo, *pgxpool.Pool) {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test: needs docker")
	}
	ctx := context.Background()

	ctr, err := tcpostgres.Run(ctx, "postgres:18-alpine",
		tcpostgres.WithDatabase("billing"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("terminate container: %v", err)
		}
	})

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	pool, err := NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}
	t.Cleanup(pool.Close)

	db := stdlib.OpenDBFromPool(pool)
	t.Cleanup(func() { _ = db.Close() })
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	if err := goose.UpContext(ctx, db, "../../../migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return NewAccountRepo(pool), pool
}

func TestWithdrawConcurrent(t *testing.T) {
	repo, pool := newTestRepo(t)
	ctx := context.Background()

	acc, err := repo.Create(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Deposit(ctx, acc.ID, 500); err != nil {
		t.Fatal(err)
	}

	const workers = 100
	var (
		wg       sync.WaitGroup
		start    = make(chan struct{})
		ok, poor atomic.Int64
		other    = make(chan error, workers)
	)

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start // все горутины ждут здесь и стартуют одновременно
			_, err := repo.Withdraw(ctx, acc.ID, 10)
			switch {
			case err == nil:
				ok.Add(1)
			case errors.Is(err, domain.ErrInsufficientFunds):
				poor.Add(1)
			default:
				other <- err
			}
		}()
	}
	close(start) // «стартовый пистолет»
	wg.Wait()
	close(other)

	for err := range other {
		t.Errorf("unexpected error: %v", err)
	}
	if ok.Load() != 50 || poor.Load() != 50 {
		t.Fatalf("ok=%d insufficient=%d, want 50/50", ok.Load(), poor.Load())
	}

	got, err := repo.Get(ctx, acc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Balance != 0 {
		t.Fatalf("balance = %d, want 0", got.Balance)
	}

	var n int
	err = pool.QueryRow(ctx,
		`SELECT count(*) FROM transactions WHERE account_id = $1 AND type = 'withdraw'`, acc.ID,
	).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n != 50 {
		t.Fatalf("withdraw records = %d, want 50", n)
	}
}
