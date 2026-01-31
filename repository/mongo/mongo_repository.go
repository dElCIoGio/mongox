package mongorepo

import (
	"context"
	"errors"
	"time"

	"github.com/dElCIoGio/mongox/document"
	"github.com/dElCIoGio/mongox/repository"
	mongospec "github.com/dElCIoGio/mongox/spec"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mopt "go.mongodb.org/mongo-driver/mongo/options"
)

// Re-export common errors for convenience.
var (
	ErrNotFound     = repository.ErrNotFound
	ErrDuplicateKey = repository.ErrDuplicateKey
)

// isDuplicateKeyError checks if the error is a MongoDB duplicate key error.
func isDuplicateKeyError(err error) bool {
	var writeErr mongo.WriteException
	if errors.As(err, &writeErr) {
		for _, we := range writeErr.WriteErrors {
			// MongoDB error code 11000 is duplicate key error
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}

type MongoRepository[T any] struct {
	coll *mongo.Collection
}

func New[T any](coll *mongo.Collection) *MongoRepository[T] {
	return &MongoRepository[T]{coll: coll}
}

// NewWithIndexes creates a new MongoRepository and ensures indexes are created.
// This constructor is useful for types that implement the document.Indexed interface.
// If index creation fails, it returns an error.
//
// Example:
//
//	type User struct {
//	    document.Base `bson:",inline"`
//	    Email string `bson:"email"`
//	}
//
//	func (u User) Indexes() []document.Index {
//	    return []document.Index{
//	        {Keys: bson.D{{"email", 1}}, Unique: true},
//	    }
//	}
//
//	repo, err := mongorepo.NewWithIndexes[User](ctx, coll)
func NewWithIndexes[T document.Indexed](ctx context.Context, coll *mongo.Collection) (*MongoRepository[T], error) {
	repo := &MongoRepository[T]{coll: coll}
	if err := repo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

// EnsureIndexes creates indexes defined by the document type's Indexes() method.
// This is automatically called by NewWithIndexes, but can also be called manually.
// If the type T does not implement document.Indexed, this method does nothing.
func (r *MongoRepository[T]) EnsureIndexes(ctx context.Context) error {
	var zero T
	indexed, ok := any(zero).(document.Indexed)
	if !ok {
		return nil
	}

	indexes := indexed.Indexes()
	if len(indexes) == 0 {
		return nil
	}

	models := make([]mongo.IndexModel, len(indexes))
	for i, idx := range indexes {
		opts := mopt.Index()
		if idx.Unique {
			opts.SetUnique(true)
		}
		if idx.Sparse {
			opts.SetSparse(true)
		}
		if idx.Name != "" {
			opts.SetName(idx.Name)
		}
		if idx.TTL != nil {
			opts.SetExpireAfterSeconds(int32(idx.TTL.Seconds()))
		}
		if idx.Background {
			opts.SetBackground(true)
		}
		if idx.PartialFilterExpression != nil {
			opts.SetPartialFilterExpression(idx.PartialFilterExpression)
		}

		models[i] = mongo.IndexModel{
			Keys:    idx.Keys,
			Options: opts,
		}
	}

	_, err := r.coll.Indexes().CreateMany(ctx, models)
	return err
}

// Collection returns the underlying mongo.Collection.
// Use this for advanced operations not covered by the repository interface.
func (r *MongoRepository[T]) Collection() *mongo.Collection {
	return r.coll
}

// ---- auto-touch helpers ----

type insertToucher interface{ TouchForInsert(time.Time) }
type updateToucher interface{ TouchForUpdate(time.Time) }

func nowUTC() time.Time { return time.Now().UTC() }

// Best-effort: if update contains a $set document, inject updated_at.
func injectUpdatedAt(update any, ts time.Time) any {
	if update == nil {
		return update
	}

	switch u := update.(type) {
	case bson.M:
		setRaw, ok := u["$set"]
		if !ok {
			return update
		}
		switch setDoc := setRaw.(type) {
		case bson.M:
			setDoc["updated_at"] = ts
			u["$set"] = setDoc
			return u
		case map[string]any:
			setDoc["updated_at"] = ts
			u["$set"] = setDoc
			return u
		default:
			return update
		}

	case bson.D:
		for i := range u {
			if u[i].Key != "$set" {
				continue
			}
			switch setDoc := u[i].Value.(type) {
			case bson.M:
				setDoc["updated_at"] = ts
				u[i].Value = setDoc
				return u
			case bson.D:
				setDoc = append(setDoc, bson.E{Key: "updated_at", Value: ts})
				u[i].Value = setDoc
				return u
			default:
				return update
			}
		}
		return update

	default:
		return update
	}
}

// ---- CRUD ----

func (r *MongoRepository[T]) InsertOne(ctx context.Context, doc *T) error {
	if doc == nil {
		return repository.ErrNilDocument
	}

	// Auto-touch if embedded Base exists (promoted methods).
	if t, ok := any(doc).(insertToucher); ok {
		t.TouchForInsert(nowUTC())
	}

	// Validate if the document implements Validatable.
	if v, ok := any(doc).(document.Validatable); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	// BeforeSave hook.
	if h, ok := any(doc).(document.BeforeSave); ok {
		if err := h.BeforeSave(ctx); err != nil {
			return err
		}
	}

	_, err := r.coll.InsertOne(ctx, doc)
	if err != nil {
		if isDuplicateKeyError(err) {
			return repository.ErrDuplicateKey
		}
		return err
	}
	return nil
}

func (r *MongoRepository[T]) FindOne(ctx context.Context, filter any, opts ...repository.FindOption) (*T, error) {
	f, err := normalizeFilter(filter)
	if err != nil {
		return nil, err
	}

	fo := applyFindOptions(opts)
	mongoOpts := mopt.FindOne()
	if fo.Sort != nil {
		mongoOpts.SetSort(fo.Sort)
	}

	var out T
	err = r.coll.FindOne(ctx, f, mongoOpts).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// AfterLoad hook.
	if h, ok := any(&out).(document.AfterLoad); ok {
		if err := h.AfterLoad(ctx); err != nil {
			return nil, err
		}
	}

	return &out, nil
}

func (r *MongoRepository[T]) Find(ctx context.Context, filter any, opts ...repository.FindOption) ([]T, error) {
	f, err := normalizeFilter(filter)
	if err != nil {
		return nil, err
	}

	fo := applyFindOptions(opts)
	mongoOpts := mopt.Find()
	if fo.Limit > 0 {
		mongoOpts.SetLimit(fo.Limit)
	}
	if fo.Skip > 0 {
		mongoOpts.SetSkip(fo.Skip)
	}
	if fo.Sort != nil {
		mongoOpts.SetSort(fo.Sort)
	}

	cur, err := r.coll.Find(ctx, f, mongoOpts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []T
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}

	// AfterLoad hook for each document (best-effort).
	for i := range results {
		if h, ok := any(&results[i]).(document.AfterLoad); ok {
			if err := h.AfterLoad(ctx); err != nil {
				return nil, err
			}
		}
	}

	return results, nil
}

// FindPaginated finds documents matching the filter with pagination.
// Returns a Page containing the documents and pagination metadata.
func (r *MongoRepository[T]) FindPaginated(ctx context.Context, filter any, page, perPage int, opts ...repository.FindOption) (*repository.Page[T], error) {
	// Normalize pagination options
	pagOpts := repository.PaginationOptions{
		Page:    page,
		PerPage: perPage,
	}
	pagOpts.Normalize()

	// Get total count
	total, err := r.Count(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	totalPages := repository.CalculateTotalPages(total, pagOpts.PerPage)

	// Build find options with pagination
	findOpts := make([]repository.FindOption, 0, len(opts)+2)
	findOpts = append(findOpts, opts...)
	findOpts = append(findOpts,
		repository.WithSkip(pagOpts.Skip()),
		repository.WithLimit(pagOpts.Limit()),
	)

	// Fetch items
	items, err := r.Find(ctx, filter, findOpts...)
	if err != nil {
		return nil, err
	}

	return &repository.Page[T]{
		Items:      items,
		Total:      total,
		Page:       pagOpts.Page,
		PerPage:    pagOpts.PerPage,
		TotalPages: totalPages,
		HasNext:    pagOpts.Page < totalPages,
		HasPrev:    pagOpts.Page > 1,
	}, nil
}

func (r *MongoRepository[T]) UpdateOne(ctx context.Context, filter any, update any) (matched int64, modified int64, err error) {
	f, err := normalizeFilter(filter)
	if err != nil {
		return 0, 0, err
	}
	if update == nil {
		return 0, 0, repository.ErrNilUpdate
	}

	// Normalize update if it implements the Update interface
	u := normalizeUpdate(update)

	// Best-effort: add updated_at to $set updates.
	u = injectUpdatedAt(u, nowUTC())

	res, err := r.coll.UpdateOne(ctx, f, u)
	if err != nil {
		return 0, 0, err
	}
	return res.MatchedCount, res.ModifiedCount, nil
}

func (r *MongoRepository[T]) DeleteOne(ctx context.Context, filter any) (deleted int64, err error) {
	f, err := normalizeFilter(filter)
	if err != nil {
		return 0, err
	}

	res, err := r.coll.DeleteOne(ctx, f)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

// ReplaceOne is useful when you want auto-touch + BeforeSave for updates.
// (Mongo UpdateOne can't mutate a doc instance, so ReplaceOne is the "document-aware" update.)
func (r *MongoRepository[T]) ReplaceOne(ctx context.Context, filter any, doc *T) (matched int64, modified int64, err error) {
	if doc == nil {
		return 0, 0, repository.ErrNilDocument
	}

	f, err := normalizeFilter(filter)
	if err != nil {
		return 0, 0, err
	}

	// Auto-touch on replace (UpdatedAt).
	if t, ok := any(doc).(updateToucher); ok {
		t.TouchForUpdate(nowUTC())
	}

	// Validate if the document implements Validatable.
	if v, ok := any(doc).(document.Validatable); ok {
		if err := v.Validate(); err != nil {
			return 0, 0, err
		}
	}

	// BeforeSave hook.
	if h, ok := any(doc).(document.BeforeSave); ok {
		if err := h.BeforeSave(ctx); err != nil {
			return 0, 0, err
		}
	}

	res, err := r.coll.ReplaceOne(ctx, f, doc)
	if err != nil {
		return 0, 0, err
	}
	return res.MatchedCount, res.ModifiedCount, nil
}

// ---- Bulk Operations ----

// InsertMany inserts multiple documents into the collection.
// Returns the ObjectIDs of the inserted documents.
func (r *MongoRepository[T]) InsertMany(ctx context.Context, docs []*T) ([]primitive.ObjectID, error) {
	if len(docs) == 0 {
		return []primitive.ObjectID{}, nil
	}

	// Prepare documents: auto-touch, validate, and call BeforeSave hooks
	now := nowUTC()
	insertDocs := make([]any, len(docs))
	for i, doc := range docs {
		if doc == nil {
			return nil, repository.ErrNilDocument
		}

		// Auto-touch if embedded Base exists
		if t, ok := any(doc).(insertToucher); ok {
			t.TouchForInsert(now)
		}

		// Validate if the document implements Validatable
		if v, ok := any(doc).(document.Validatable); ok {
			if err := v.Validate(); err != nil {
				return nil, err
			}
		}

		// BeforeSave hook
		if h, ok := any(doc).(document.BeforeSave); ok {
			if err := h.BeforeSave(ctx); err != nil {
				return nil, err
			}
		}

		insertDocs[i] = doc
	}

	res, err := r.coll.InsertMany(ctx, insertDocs)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, repository.ErrDuplicateKey
		}
		return nil, err
	}

	ids := make([]primitive.ObjectID, len(res.InsertedIDs))
	for i, id := range res.InsertedIDs {
		if oid, ok := id.(primitive.ObjectID); ok {
			ids[i] = oid
		}
	}

	return ids, nil
}

// UpdateMany updates all documents matching the filter.
// Returns the number of documents matched and modified.
func (r *MongoRepository[T]) UpdateMany(ctx context.Context, filter any, update any) (matched int64, modified int64, err error) {
	f, err := normalizeFilter(filter)
	if err != nil {
		return 0, 0, err
	}
	if update == nil {
		return 0, 0, repository.ErrNilUpdate
	}

	// Normalize update if it implements the Update interface
	u := normalizeUpdate(update)

	// Best-effort: add updated_at to $set updates
	u = injectUpdatedAt(u, nowUTC())

	res, err := r.coll.UpdateMany(ctx, f, u)
	if err != nil {
		return 0, 0, err
	}
	return res.MatchedCount, res.ModifiedCount, nil
}

// DeleteMany deletes all documents matching the filter.
// Returns the number of documents deleted.
func (r *MongoRepository[T]) DeleteMany(ctx context.Context, filter any) (deleted int64, err error) {
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

// Count returns the number of documents matching the filter.
func (r *MongoRepository[T]) Count(ctx context.Context, filter any) (int64, error) {
	f, err := normalizeFilter(filter)
	if err != nil {
		return 0, err
	}

	return r.coll.CountDocuments(ctx, f)
}

// BulkWrite executes multiple write operations in a single batch.
// Returns a BulkWriteResult with counts of affected documents.
func (r *MongoRepository[T]) BulkWrite(ctx context.Context, ops []repository.BulkOp) (*repository.BulkWriteResult, error) {
	if len(ops) == 0 {
		return &repository.BulkWriteResult{}, nil
	}

	models := make([]mongo.WriteModel, 0, len(ops))

	for _, op := range ops {
		switch op.Type {
		case repository.BulkOpInsert:
			models = append(models, mongo.NewInsertOneModel().SetDocument(op.Doc))

		case repository.BulkOpUpdate:
			f, err := normalizeFilter(op.Filter)
			if err != nil {
				return nil, err
			}
			u := normalizeUpdate(op.Update)
			model := mongo.NewUpdateOneModel().SetFilter(f).SetUpdate(u).SetUpsert(op.Upsert)
			models = append(models, model)

		case repository.BulkOpReplace:
			f, err := normalizeFilter(op.Filter)
			if err != nil {
				return nil, err
			}
			model := mongo.NewReplaceOneModel().SetFilter(f).SetReplacement(op.Doc).SetUpsert(op.Upsert)
			models = append(models, model)

		case repository.BulkOpDelete:
			f, err := normalizeFilter(op.Filter)
			if err != nil {
				return nil, err
			}
			models = append(models, mongo.NewDeleteOneModel().SetFilter(f))
		}
	}

	res, err := r.coll.BulkWrite(ctx, models)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, repository.ErrDuplicateKey
		}
		return nil, err
	}

	upsertedIDs := make(map[int64]primitive.ObjectID)
	for idx, id := range res.UpsertedIDs {
		if oid, ok := id.(primitive.ObjectID); ok {
			upsertedIDs[idx] = oid
		}
	}

	return &repository.BulkWriteResult{
		InsertedCount: res.InsertedCount,
		MatchedCount:  res.MatchedCount,
		ModifiedCount: res.ModifiedCount,
		DeletedCount:  res.DeletedCount,
		UpsertedCount: res.UpsertedCount,
		UpsertedIDs:   upsertedIDs,
	}, nil
}

// ---- Aggregation ----

// pipelineConverter is implemented by types that can be converted to a MongoDB pipeline.
type pipelineConverter interface {
	ToPipeline() []bson.M
}

func normalizePipeline(pipeline any) ([]bson.M, error) {
	if pipeline == nil {
		return []bson.M{}, nil
	}

	switch p := pipeline.(type) {
	case []bson.M:
		return p, nil
	case []bson.D:
		result := make([]bson.M, len(p))
		for i, stage := range p {
			result[i] = bson.M{}
			for _, elem := range stage {
				result[i][elem.Key] = elem.Value
			}
		}
		return result, nil
	case pipelineConverter:
		return p.ToPipeline(), nil
	default:
		return nil, repository.ErrInvalidFilter
	}
}

// Aggregate executes an aggregation pipeline and returns the results decoded as type T.
// The pipeline can be []bson.M, []bson.D, or a Pipeline builder.
func (r *MongoRepository[T]) Aggregate(ctx context.Context, pipeline any) ([]T, error) {
	p, err := normalizePipeline(pipeline)
	if err != nil {
		return nil, err
	}

	cur, err := r.coll.Aggregate(ctx, p)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []T
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}

	// AfterLoad hook for each document.
	for i := range results {
		if h, ok := any(&results[i]).(document.AfterLoad); ok {
			if err := h.AfterLoad(ctx); err != nil {
				return nil, err
			}
		}
	}

	return results, nil
}

// AggregateRaw executes an aggregation pipeline and returns raw bson.M results.
// Use this when the aggregation output doesn't match type T.
func (r *MongoRepository[T]) AggregateRaw(ctx context.Context, pipeline any) ([]bson.M, error) {
	p, err := normalizePipeline(pipeline)
	if err != nil {
		return nil, err
	}

	cur, err := r.coll.Aggregate(ctx, p)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []bson.M
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// ---- helpers ----

func normalizeFilter(filter any) (any, error) {
	if filter == nil {
		return bson.M{}, nil
	}
	if f, ok := filter.(mongospec.Filter); ok {
		return f.ToMongo(), nil
	}
	return filter, nil
}

// updateConverter is implemented by types that can be converted to a MongoDB update.
type updateConverter interface {
	ToBsonUpdate() bson.M
}

func normalizeUpdate(update any) any {
	if update == nil {
		return update
	}
	if u, ok := update.(updateConverter); ok {
		return u.ToBsonUpdate()
	}
	return update
}

func applyFindOptions(opts []repository.FindOption) repository.FindOptions {
	var fo repository.FindOptions
	for _, fn := range opts {
		if fn != nil {
			fn(&fo)
		}
	}
	return fo
}
