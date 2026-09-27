// Package repository provides explicit pgx execution primitives and generic
// persistence operations. Modules own their schema metadata, scanners, write
// values, and domain-specific predicates; this package never reflects over
// entities or accepts request-provided SQL.
package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is the common pgx execution surface implemented by a pool and a
// transaction. Repositories receive this narrow dependency so callers can
// explicitly rebind them to a transaction.
type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// RowScanner is implemented by pgx.Row and pgx.Rows.
type RowScanner interface {
	Scan(...any) error
}

// CommandResult records the observed affected-row count for a mutation.
type CommandResult struct {
	RowsAffected int64
}

// Binding stores a repository execution dependency and produces a copied
// binding for a transaction-bound DBTX.
type Binding struct {
	db DBTX
}

// NewBinding creates an execution binding for a pool or transaction.
func NewBinding(db DBTX) Binding {
	return Binding{db: db}
}

// WithDB returns a copy bound to db without mutating the original binding.
func (value Binding) WithDB(db DBTX) Binding {
	return Binding{db: db}
}

// DB returns the pool- or transaction-bound execution dependency.
func (value Binding) DB() DBTX {
	return value.db
}

func commandResult(tag pgconn.CommandTag) CommandResult {
	return CommandResult{RowsAffected: tag.RowsAffected()}
}
