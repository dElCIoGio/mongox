package mongorepo

import (
	"context"
	"errors"

	"github.com/dElCIoGio/mongox/repository"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// MongoTransactionManager implements TransactionManager using MongoDB sessions.
type MongoTransactionManager struct {
	client *mongo.Client
	opts   *repository.TransactionOptions
}

// NewTransactionManager creates a new TransactionManager for the given MongoDB client.
func NewTransactionManager(client *mongo.Client, opts *repository.TransactionOptions) *MongoTransactionManager {
	return &MongoTransactionManager{
		client: client,
		opts:   opts,
	}
}

// WithTransaction executes the given function within a MongoDB transaction.
// The function receives a context that should be used for all database operations.
// If the function returns an error, the transaction is aborted.
// If the function returns nil, the transaction is committed.
func (tm *MongoTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Start a session
	session, err := tm.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Build transaction options
	txnOpts := options.Transaction()
	if tm.opts != nil {
		if tm.opts.ReadConcern != "" {
			txnOpts.SetReadConcern(parseReadConcern(tm.opts.ReadConcern))
		}
		if tm.opts.WriteConcern != nil {
			txnOpts.SetWriteConcern(parseWriteConcern(tm.opts.WriteConcern))
		}
	}

	// Execute the transaction with automatic retry for transient errors
	maxRetries := 0
	if tm.opts != nil {
		maxRetries = tm.opts.MaxRetries
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		lastErr = mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
			// Start the transaction
			if err := session.StartTransaction(txnOpts); err != nil {
				return err
			}

			// Execute the user function
			if err := fn(sessCtx); err != nil {
				// Abort on error
				_ = session.AbortTransaction(sessCtx)
				return err
			}

			// Commit the transaction
			return session.CommitTransaction(sessCtx)
		})

		if lastErr == nil {
			return nil
		}

		// Check if error is retryable
		if !isTransientTransactionError(lastErr) {
			return lastErr
		}
	}

	return lastErr
}

// isTransientTransactionError checks if the error is a transient transaction error
// that can be retried.
func isTransientTransactionError(err error) bool {
	var cmdErr mongo.CommandError
	if errors.As(err, &cmdErr) {
		return cmdErr.HasErrorLabel("TransientTransactionError")
	}
	return false
}

func parseReadConcern(level string) *readconcern.ReadConcern {
	switch level {
	case "local":
		return readconcern.Local()
	case "available":
		return readconcern.Available()
	case "majority":
		return readconcern.Majority()
	case "linearizable":
		return readconcern.Linearizable()
	case "snapshot":
		return readconcern.Snapshot()
	default:
		return readconcern.Local()
	}
}

func parseWriteConcern(wc *repository.WriteConcern) *writeconcern.WriteConcern {
	var opts []writeconcern.Option

	if wc.W != nil {
		switch w := wc.W.(type) {
		case int:
			opts = append(opts, writeconcern.W(w))
		case string:
			if w == "majority" {
				opts = append(opts, writeconcern.WMajority())
			}
		}
	}

	if wc.Journal {
		opts = append(opts, writeconcern.J(true))
	}

	if wc.WTimeout > 0 {
		opts = append(opts, writeconcern.WTimeout(wc.WTimeout))
	}

	return writeconcern.New(opts...)
}

// RunInTransaction is a convenience function that creates a transaction manager
// and executes the function within a transaction.
func RunInTransaction(ctx context.Context, client *mongo.Client, fn func(ctx context.Context) error) error {
	tm := NewTransactionManager(client, nil)
	return tm.WithTransaction(ctx, fn)
}

// RunInTransactionWithRetry is like RunInTransaction but with automatic retry
// on transient errors.
func RunInTransactionWithRetry(ctx context.Context, client *mongo.Client, maxRetries int, fn func(ctx context.Context) error) error {
	tm := NewTransactionManager(client, &repository.TransactionOptions{
		MaxRetries: maxRetries,
	})
	return tm.WithTransaction(ctx, fn)
}
