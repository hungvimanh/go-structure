package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TransactionCallback executes with a transaction-bound pgx.Tx.
type TransactionCallback func(context.Context, pgx.Tx) error

// Transactor starts top-level pgx transactions. It intentionally provides no
// nested-transaction guarantee; callers that need savepoints must own that
// behavior explicitly.
type Transactor struct {
	beginner interface {
		Begin(context.Context) (pgx.Tx, error)
	}
}

// NewTransactor creates a transaction boundary over a pgx pool.
func NewTransactor(pool *pgxpool.Pool) *Transactor {
	return &Transactor{beginner: pool}
}

// WithinTx executes callback atomically. Callback errors are returned without
// wrapping so context cancellation and deadline causes remain intact. Rollback
// is best-effort; a rollback infrastructure error is joined with the callback
// error without hiding either cause.
func (value *Transactor) WithinTx(ctx context.Context, callback TransactionCallback) error {
	transaction, err := value.beginner.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = transaction.Rollback(context.WithoutCancel(ctx))
			panic(recovered)
		}
	}()

	if err := callback(ctx, transaction); err != nil {
		if rollbackErr := transaction.Rollback(context.WithoutCancel(ctx)); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			return errors.Join(err, fmt.Errorf("rollback transaction: %w", rollbackErr))
		}
		return err
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
