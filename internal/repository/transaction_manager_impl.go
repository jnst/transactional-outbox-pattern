package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TransactionManagerImpl implements TransactionManager using PostgreSQL.
type TransactionManagerImpl struct {
	pool *pgxpool.Pool
}

// NewTransactionManagerImpl creates a new TransactionManager implementation.
func NewTransactionManagerImpl(pool *pgxpool.Pool) TransactionManager {
	return &TransactionManagerImpl{pool: pool}
}

// WithTransaction executes a function within a database transaction.
func (tm *TransactionManagerImpl) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(ctx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("transaction failed: %w, rollback failed: %v", err, rollbackErr)
		}

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("commit failed: %w, rollback failed: %v", err, rollbackErr)
		}

		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
