package spec

import "go.mongodb.org/mongo-driver/bson"

type andFilter struct {
	filters []Filter
}

func (f andFilter) ToMongo() bson.M {
	parts := make([]bson.M, 0, len(f.filters))
	for _, flt := range f.filters {
		if flt == nil {
			continue
		}
		parts = append(parts, flt.ToMongo())
	}

	// If there's only one filter, return it directly (avoids unnecessary $and nesting).
	if len(parts) == 1 {
		return parts[0]
	}

	return bson.M{"$and": parts}
}

// And combines multiple filters with a logical AND operation.
// All conditions must be true for a document to match.
//
// Behavior:
//   - Nil filters are automatically ignored
//   - If only one non-nil filter is provided, it returns that filter directly (no wrapping)
//   - Nested And() calls are automatically flattened for cleaner queries
//   - Returns nil if all filters are nil
//
// MongoDB equivalent: {$and: [filter1, filter2, ...]}
//
// Example:
//
//	// Match active users over 18
//	And(Eq("status", "active"), Gte("age", 18))
//	// MongoDB: {"$and": [{"status": "active"}, {"age": {"$gte": 18}}]}
//
//	// Complex nested conditions (automatically flattened)
//	And(
//	    And(Eq("a", 1), Eq("b", 2)),
//	    Eq("c", 3),
//	)
//	// MongoDB: {"$and": [{"a": 1}, {"b": 2}, {"c": 3}]}
func And(filters ...Filter) Filter {
	flat := make([]Filter, 0, len(filters))
	for _, f := range filters {
		if f == nil {
			continue
		}
		// Flatten nested And(...) to keep filters tidy.
		if af, ok := f.(andFilter); ok {
			flat = append(flat, af.filters...)
			continue
		}
		flat = append(flat, f)
	}

	if len(flat) == 0 {
		return nil
	}
	if len(flat) == 1 {
		return flat[0]
	}
	return andFilter{filters: flat}
}

type orFilter struct {
	filters []Filter
}

func (f orFilter) ToMongo() bson.M {
	parts := make([]bson.M, 0, len(f.filters))
	for _, flt := range f.filters {
		if flt == nil {
			continue
		}
		parts = append(parts, flt.ToMongo())
	}

	// If there's only one filter, return it directly (avoids unnecessary $or nesting).
	if len(parts) == 1 {
		return parts[0]
	}

	return bson.M{"$or": parts}
}

// Or combines multiple filters with a logical OR operation.
// At least one condition must be true for a document to match.
//
// Behavior:
//   - Nil filters are automatically ignored
//   - If only one non-nil filter is provided, it returns that filter directly (no wrapping)
//   - Nested Or() calls are automatically flattened for cleaner queries
//   - Returns nil if all filters are nil
//
// MongoDB equivalent: {$or: [filter1, filter2, ...]}
//
// Example:
//
//	// Match users who are either admins or have premium status
//	Or(Eq("role", "admin"), Eq("premium", true))
//	// MongoDB: {"$or": [{"role": "admin"}, {"premium": true}]}
//
//	// Match by multiple possible statuses
//	Or(Eq("status", "pending"), Eq("status", "review"), Eq("status", "approved"))
func Or(filters ...Filter) Filter {
	flat := make([]Filter, 0, len(filters))
	for _, f := range filters {
		if f == nil {
			continue
		}
		// Flatten nested Or(...) to keep filters tidy.
		if of, ok := f.(orFilter); ok {
			flat = append(flat, of.filters...)
			continue
		}
		flat = append(flat, f)
	}

	if len(flat) == 0 {
		return nil
	}
	if len(flat) == 1 {
		return flat[0]
	}
	return orFilter{filters: flat}
}

type notFilter struct {
	filter Filter
}

func (f notFilter) ToMongo() bson.M {
	// MongoDB: { field: { $not: { $gt: ... } } } only works for certain operators,
	// but using $nor is a reliable general NOT for full filters.
	return bson.M{"$nor": []bson.M{f.filter.ToMongo()}}
}

// Not negates a filter using MongoDB's $nor operator for general-purpose negation.
// Documents that do NOT match the filter will be returned.
//
// Note: This uses $nor internally because MongoDB's $not operator only works
// with specific field-level operators and cannot negate arbitrary filter expressions.
//
// Behavior:
//   - Returns nil if the input filter is nil
//
// MongoDB equivalent: {$nor: [filter]}
//
// Example:
//
//	// Match users who are NOT admins
//	Not(Eq("role", "admin"))
//	// MongoDB: {"$nor": [{"role": "admin"}]}
//
//	// Match orders that are NOT (pending AND low priority)
//	Not(And(Eq("status", "pending"), Lt("priority", 3)))
func Not(filter Filter) Filter {
	if filter == nil {
		return nil
	}
	return notFilter{filter: filter}
}
