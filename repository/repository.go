package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Repository is a minimal CRUD interface for a collection of T.
type Repository[T any] interface {
	// Single document operations
	InsertOne(ctx context.Context, doc *T) error
	FindOne(ctx context.Context, filter any, opts ...FindOption) (*T, error)
	Find(ctx context.Context, filter any, opts ...FindOption) ([]T, error)
	UpdateOne(ctx context.Context, filter any, update any) (matched int64, modified int64, err error)
	ReplaceOne(ctx context.Context, filter any, doc *T) (matched int64, modified int64, err error)
	DeleteOne(ctx context.Context, filter any) (deleted int64, err error)

	// Bulk operations
	InsertMany(ctx context.Context, docs []*T) ([]primitive.ObjectID, error)
	UpdateMany(ctx context.Context, filter any, update any) (matched int64, modified int64, err error)
	DeleteMany(ctx context.Context, filter any) (deleted int64, err error)

	// Aggregate executes an aggregation pipeline and returns the results.
	// The pipeline can be []bson.M, []bson.D, or a Pipeline builder.
	Aggregate(ctx context.Context, pipeline any) ([]T, error)

	// AggregateRaw executes an aggregation pipeline and returns raw bson.M results.
	// Use this when the aggregation output doesn't match type T.
	AggregateRaw(ctx context.Context, pipeline any) ([]bson.M, error)

	// Count returns the number of documents matching the filter.
	Count(ctx context.Context, filter any) (int64, error)
}

// BulkWriteResult contains the results of a bulk write operation.
type BulkWriteResult struct {
	InsertedCount int64
	MatchedCount  int64
	ModifiedCount int64
	DeletedCount  int64
	UpsertedCount int64
	UpsertedIDs   map[int64]primitive.ObjectID
}
