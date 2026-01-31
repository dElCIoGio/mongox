package repository

// FindOption is a functional option for configuring Find and FindOne operations.
// Use the With* functions to create options.
//
// Example:
//
//	users, err := repo.Find(ctx, filter,
//	    WithLimit(10),
//	    WithSkip(20),
//	    WithSort(bson.D{{"created_at", -1}}),
//	)
type FindOption func(*FindOptions)

// FindOptions contains the configuration for Find operations.
// This struct is populated by applying FindOption functions.
type FindOptions struct {
	// Limit specifies the maximum number of documents to return.
	// A value of 0 means no limit.
	Limit int64

	// Skip specifies the number of documents to skip before returning results.
	// Useful for pagination.
	Skip int64

	// Sort specifies the order in which to return documents.
	// Typically bson.D for ordered sorting, e.g., bson.D{{"created_at", -1}}.
	Sort any
}

// WithLimit creates an option that limits the number of documents returned.
// Pass 0 to remove any limit.
//
// Example:
//
//	WithLimit(10)    // Return at most 10 documents
//	WithLimit(0)     // No limit (return all matching documents)
func WithLimit(limit int64) FindOption {
	return func(o *FindOptions) { o.Limit = limit }
}

// WithSkip creates an option that skips the specified number of documents.
// Commonly used with WithLimit for pagination.
//
// Example:
//
//	// Page 2 with 10 items per page
//	WithSkip(10), WithLimit(10)
func WithSkip(skip int64) FindOption {
	return func(o *FindOptions) { o.Skip = skip }
}

// WithSort creates an option that specifies the sort order for results.
// The sort parameter should be a bson.D for predictable ordering.
//
// Sort order values:
//   - 1: Ascending order
//   - -1: Descending order
//
// Example:
//
//	// Sort by created_at descending (newest first)
//	WithSort(bson.D{{"created_at", -1}})
//
//	// Sort by multiple fields
//	WithSort(bson.D{{"status", 1}, {"priority", -1}})
func WithSort(sort any) FindOption {
	return func(o *FindOptions) { o.Sort = sort }
}

// applyFindOptions applies all provided options to create a FindOptions struct.
func applyFindOptions(opts []FindOption) FindOptions {
	var o FindOptions
	for _, fn := range opts {
		if fn != nil {
			fn(&o)
		}
	}
	return o
}
