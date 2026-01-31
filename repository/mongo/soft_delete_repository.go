package mongorepo

import (
	"context"
	"time"

	"github.com/dElCIoGio/mongox/repository"
	mongospec "github.com/dElCIoGio/mongox/spec"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// SoftDeleteRepository extends MongoRepository with soft delete functionality.
// Documents are marked as deleted instead of being removed from the database.
// Find operations automatically exclude deleted documents unless FindWithDeleted is used.
type SoftDeleteRepository[T any] struct {
	*MongoRepository[T]
}

// NewSoftDelete creates a new SoftDeleteRepository wrapping the given collection.
func NewSoftDelete[T any](coll *mongo.Collection) *SoftDeleteRepository[T] {
	return &SoftDeleteRepository[T]{
		MongoRepository: New[T](coll),
	}
}

// notDeletedFilter returns a filter that excludes soft-deleted documents.
func notDeletedFilter() bson.M {
	return bson.M{"deleted_at": bson.M{"$exists": false}}
}

// combineWithNotDeleted combines the given filter with the not-deleted filter.
func combineWithNotDeleted(filter any) any {
	notDeleted := notDeletedFilter()

	if filter == nil {
		return notDeleted
	}

	// Handle Filter interface
	if f, ok := filter.(mongospec.Filter); ok {
		return bson.M{"$and": []bson.M{f.ToMongo(), notDeleted}}
	}

	// Handle bson.M
	if m, ok := filter.(bson.M); ok {
		return bson.M{"$and": []bson.M{m, notDeleted}}
	}

	// Handle bson.D
	if d, ok := filter.(bson.D); ok {
		filterM := bson.M{}
		for _, e := range d {
			filterM[e.Key] = e.Value
		}
		return bson.M{"$and": []bson.M{filterM, notDeleted}}
	}

	// Fallback: wrap in $and
	return bson.M{"$and": []any{filter, notDeleted}}
}

// FindOne finds a single non-deleted document matching the filter.
func (r *SoftDeleteRepository[T]) FindOne(ctx context.Context, filter any, opts ...repository.FindOption) (*T, error) {
	return r.MongoRepository.FindOne(ctx, combineWithNotDeleted(filter), opts...)
}

// Find finds all non-deleted documents matching the filter.
func (r *SoftDeleteRepository[T]) Find(ctx context.Context, filter any, opts ...repository.FindOption) ([]T, error) {
	return r.MongoRepository.Find(ctx, combineWithNotDeleted(filter), opts...)
}

// FindWithDeleted finds documents including soft-deleted ones.
// Use this when you need to access deleted documents.
func (r *SoftDeleteRepository[T]) FindWithDeleted(ctx context.Context, filter any, opts ...repository.FindOption) ([]T, error) {
	return r.MongoRepository.Find(ctx, filter, opts...)
}

// FindOneWithDeleted finds a single document including soft-deleted ones.
func (r *SoftDeleteRepository[T]) FindOneWithDeleted(ctx context.Context, filter any, opts ...repository.FindOption) (*T, error) {
	return r.MongoRepository.FindOne(ctx, filter, opts...)
}

// FindDeleted finds only soft-deleted documents matching the filter.
func (r *SoftDeleteRepository[T]) FindDeleted(ctx context.Context, filter any, opts ...repository.FindOption) ([]T, error) {
	deletedFilter := bson.M{"deleted_at": bson.M{"$exists": true}}

	if filter == nil {
		return r.MongoRepository.Find(ctx, deletedFilter, opts...)
	}

	// Handle Filter interface
	if f, ok := filter.(mongospec.Filter); ok {
		return r.MongoRepository.Find(ctx, bson.M{"$and": []bson.M{f.ToMongo(), deletedFilter}}, opts...)
	}

	// Handle bson.M
	if m, ok := filter.(bson.M); ok {
		return r.MongoRepository.Find(ctx, bson.M{"$and": []bson.M{m, deletedFilter}}, opts...)
	}

	return r.MongoRepository.Find(ctx, bson.M{"$and": []any{filter, deletedFilter}}, opts...)
}

// SoftDelete marks documents matching the filter as deleted by setting deleted_at.
// Returns the number of documents that were soft deleted.
func (r *SoftDeleteRepository[T]) SoftDelete(ctx context.Context, filter any) (int64, error) {
	// Only soft-delete non-deleted documents
	f := combineWithNotDeleted(filter)

	update := bson.M{"$set": bson.M{"deleted_at": time.Now().UTC()}}
	matched, _, err := r.MongoRepository.UpdateOne(ctx, f, update)
	return matched, err
}

// SoftDeleteMany marks all documents matching the filter as deleted.
// Returns the number of documents that were soft deleted.
func (r *SoftDeleteRepository[T]) SoftDeleteMany(ctx context.Context, filter any) (int64, error) {
	// Only soft-delete non-deleted documents
	f := combineWithNotDeleted(filter)

	update := bson.M{"$set": bson.M{"deleted_at": time.Now().UTC()}}

	res, err := r.coll.UpdateMany(ctx, f, update)
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}

// Restore removes the deleted_at timestamp from documents matching the filter.
// This restores soft-deleted documents.
// Returns the number of documents that were restored.
func (r *SoftDeleteRepository[T]) Restore(ctx context.Context, filter any) (int64, error) {
	// Only restore deleted documents
	deletedFilter := bson.M{"deleted_at": bson.M{"$exists": true}}
	var f any

	if filter == nil {
		f = deletedFilter
	} else if flt, ok := filter.(mongospec.Filter); ok {
		f = bson.M{"$and": []bson.M{flt.ToMongo(), deletedFilter}}
	} else if m, ok := filter.(bson.M); ok {
		f = bson.M{"$and": []bson.M{m, deletedFilter}}
	} else {
		f = bson.M{"$and": []any{filter, deletedFilter}}
	}

	update := bson.M{"$unset": bson.M{"deleted_at": ""}}
	matched, _, err := r.MongoRepository.UpdateOne(ctx, f, update)
	return matched, err
}

// RestoreMany restores all soft-deleted documents matching the filter.
// Returns the number of documents that were restored.
func (r *SoftDeleteRepository[T]) RestoreMany(ctx context.Context, filter any) (int64, error) {
	deletedFilter := bson.M{"deleted_at": bson.M{"$exists": true}}
	var f any

	if filter == nil {
		f = deletedFilter
	} else if flt, ok := filter.(mongospec.Filter); ok {
		f = bson.M{"$and": []bson.M{flt.ToMongo(), deletedFilter}}
	} else if m, ok := filter.(bson.M); ok {
		f = bson.M{"$and": []bson.M{m, deletedFilter}}
	} else {
		f = bson.M{"$and": []any{filter, deletedFilter}}
	}

	update := bson.M{"$unset": bson.M{"deleted_at": ""}}

	res, err := r.coll.UpdateMany(ctx, f, update)
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}

// HardDelete permanently removes documents matching the filter.
// Use with caution - this cannot be undone.
func (r *SoftDeleteRepository[T]) HardDelete(ctx context.Context, filter any) (int64, error) {
	return r.MongoRepository.DeleteOne(ctx, filter)
}

// HardDeleteMany permanently removes all documents matching the filter.
// Use with caution - this cannot be undone.
func (r *SoftDeleteRepository[T]) HardDeleteMany(ctx context.Context, filter any) (int64, error) {
	f, err := normalizeFilter(filter)
	if err != nil {
		return 0, err
	}

	res, err := r.coll.DeleteMany(ctx, f)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

// Purge permanently removes all soft-deleted documents matching the filter.
// This is useful for cleaning up old deleted data.
func (r *SoftDeleteRepository[T]) Purge(ctx context.Context, filter any) (int64, error) {
	deletedFilter := bson.M{"deleted_at": bson.M{"$exists": true}}
	var f any

	if filter == nil {
		f = deletedFilter
	} else if flt, ok := filter.(mongospec.Filter); ok {
		f = bson.M{"$and": []bson.M{flt.ToMongo(), deletedFilter}}
	} else if m, ok := filter.(bson.M); ok {
		f = bson.M{"$and": []bson.M{m, deletedFilter}}
	} else {
		f = bson.M{"$and": []any{filter, deletedFilter}}
	}

	res, err := r.coll.DeleteMany(ctx, f)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

// CountActive returns the count of non-deleted documents matching the filter.
func (r *SoftDeleteRepository[T]) CountActive(ctx context.Context, filter any) (int64, error) {
	f := combineWithNotDeleted(filter)

	fNorm, err := normalizeFilter(f)
	if err != nil {
		return 0, err
	}

	return r.coll.CountDocuments(ctx, fNorm)
}

// CountDeleted returns the count of soft-deleted documents matching the filter.
func (r *SoftDeleteRepository[T]) CountDeleted(ctx context.Context, filter any) (int64, error) {
	deletedFilter := bson.M{"deleted_at": bson.M{"$exists": true}}
	var f any

	if filter == nil {
		f = deletedFilter
	} else if flt, ok := filter.(mongospec.Filter); ok {
		f = bson.M{"$and": []bson.M{flt.ToMongo(), deletedFilter}}
	} else if m, ok := filter.(bson.M); ok {
		f = bson.M{"$and": []bson.M{m, deletedFilter}}
	} else {
		f = bson.M{"$and": []any{filter, deletedFilter}}
	}

	return r.coll.CountDocuments(ctx, f)
}
