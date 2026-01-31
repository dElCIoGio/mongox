package repository

import (
	"context"
	"time"
)

// TransactionManager provides methods for executing operations within a transaction.
type TransactionManager interface {
	// WithTransaction executes the given function within a transaction.
	// If the function returns an error, the transaction is aborted.
	// If the function returns nil, the transaction is committed.
	// The context passed to the function should be used for all database operations
	// to ensure they are part of the transaction.
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// TransactionOptions configures transaction behavior.
type TransactionOptions struct {
	// MaxRetries is the maximum number of times to retry the transaction on transient errors.
	// Default is 0 (no retries).
	MaxRetries int

	// ReadConcern specifies the read concern for the transaction.
	// Valid values: "local", "available", "majority", "linearizable", "snapshot"
	ReadConcern string

	// WriteConcern specifies the write concern for the transaction.
	WriteConcern *WriteConcern
}

// WriteConcern specifies the write concern level.
type WriteConcern struct {
	// W specifies the write concern. Can be a number or "majority".
	W any

	// Journal specifies whether to wait for journal commit.
	Journal bool

	// WTimeout specifies the write concern timeout.
	WTimeout time.Duration
}
