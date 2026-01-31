package spec

import "go.mongodb.org/mongo-driver/bson"

// Filter represents a MongoDB query filter that can be translated to bson.M.
// Filters are composable building blocks for constructing MongoDB queries
// in a type-safe, readable manner.
//
// Use the provided filter functions (Eq, Gt, In, etc.) to create filters,
// and combine them with logical operators (And, Or, Not).
//
// Example:
//
//	filter := And(
//	    Eq("status", "active"),
//	    Gte("age", 18),
//	)
//	// MongoDB: {"$and": [{"status": "active"}, {"age": {"$gte": 18}}]}
type Filter interface {
	// ToMongo converts the filter to a MongoDB bson.M query document.
	ToMongo() bson.M
}

type eqFilter struct {
	field string
	value any
}

func (f eqFilter) ToMongo() bson.M {
	return bson.M{f.field: f.value}
}

// Eq creates a filter that matches documents where field equals value.
// This is the most common filter operation.
//
// MongoDB equivalent: {field: value}
//
// Example:
//
//	Eq("status", "active")     // {"status": "active"}
//	Eq("count", 42)            // {"count": 42}
//	Eq("_id", objectID)        // {"_id": ObjectId("...")}
func Eq(field string, value any) Filter {
	return eqFilter{field: field, value: value}
}

type opFilter struct {
	field string
	op    string
	value any
}

func (f opFilter) ToMongo() bson.M {
	return bson.M{f.field: bson.M{f.op: f.value}}
}

// Ne creates a filter that matches documents where field does not equal value.
//
// MongoDB equivalent: {field: {$ne: value}}
//
// Example:
//
//	Ne("status", "deleted")    // {"status": {"$ne": "deleted"}}
func Ne(field string, value any) Filter {
	return opFilter{field: field, op: "$ne", value: value}
}

// Gt creates a filter that matches documents where field is greater than value.
//
// MongoDB equivalent: {field: {$gt: value}}
//
// Example:
//
//	Gt("age", 18)              // {"age": {"$gt": 18}}
//	Gt("price", 99.99)         // {"price": {"$gt": 99.99}}
func Gt(field string, value any) Filter {
	return opFilter{field: field, op: "$gt", value: value}
}

// Gte creates a filter that matches documents where field is greater than or equal to value.
//
// MongoDB equivalent: {field: {$gte: value}}
//
// Example:
//
//	Gte("age", 18)             // {"age": {"$gte": 18}}
//	Gte("created_at", time)    // {"created_at": {"$gte": ISODate(...)}}
func Gte(field string, value any) Filter {
	return opFilter{field: field, op: "$gte", value: value}
}

// Lt creates a filter that matches documents where field is less than value.
//
// MongoDB equivalent: {field: {$lt: value}}
//
// Example:
//
//	Lt("quantity", 10)         // {"quantity": {"$lt": 10}}
func Lt(field string, value any) Filter {
	return opFilter{field: field, op: "$lt", value: value}
}

// Lte creates a filter that matches documents where field is less than or equal to value.
//
// MongoDB equivalent: {field: {$lte: value}}
//
// Example:
//
//	Lte("priority", 5)         // {"priority": {"$lte": 5}}
func Lte(field string, value any) Filter {
	return opFilter{field: field, op: "$lte", value: value}
}

// In creates a filter that matches documents where field equals any value in the slice.
// The values parameter should be a slice type (e.g., []string, []int, []primitive.ObjectID).
//
// MongoDB equivalent: {field: {$in: [values...]}}
//
// Example:
//
//	In("status", []string{"pending", "active"})
//	// {"status": {"$in": ["pending", "active"]}}
//
//	In("category_id", []primitive.ObjectID{id1, id2})
//	// {"category_id": {"$in": [ObjectId("..."), ObjectId("...")]}}
func In(field string, values any) Filter {
	return opFilter{field: field, op: "$in", value: values}
}

// NotIn creates a filter that matches documents where field does not equal any value in the slice.
// The values parameter should be a slice type.
//
// MongoDB equivalent: {field: {$nin: [values...]}}
//
// Example:
//
//	NotIn("role", []string{"banned", "suspended"})
//	// {"role": {"$nin": ["banned", "suspended"]}}
func NotIn(field string, values any) Filter {
	return opFilter{field: field, op: "$nin", value: values}
}

// Exists creates a filter that matches documents based on field existence.
// When exists is true, matches documents that contain the field.
// When exists is false, matches documents that do not contain the field.
//
// MongoDB equivalent: {field: {$exists: bool}}
//
// Example:
//
//	Exists("email", true)      // {"email": {"$exists": true}}
//	Exists("deleted_at", false) // {"deleted_at": {"$exists": false}}
func Exists(field string, exists bool) Filter {
	return opFilter{field: field, op: "$exists", value: exists}
}

// Nin is an alias for NotIn. It matches documents where field is not in the values slice.
// This alias is provided for consistency with MongoDB's $nin operator naming.
//
// MongoDB equivalent: {field: {$nin: [values...]}}
func Nin(field string, values any) Filter {
	return NotIn(field, values)
}

// Regex creates a filter that matches documents where field matches the regular expression pattern.
// Optional regex options can be provided (e.g., "i" for case-insensitive, "m" for multiline).
//
// MongoDB equivalent: {field: {$regex: pattern, $options: options}}
//
// Common options:
//   - "i": Case-insensitive matching
//   - "m": Multiline matching (^ and $ match line boundaries)
//   - "s": Dot (.) matches newlines
//   - "x": Extended mode (ignore whitespace and comments)
//
// Example:
//
//	Regex("email", "@gmail\\.com$")           // Emails ending with @gmail.com
//	Regex("name", "^john", "i")               // Names starting with "john" (case-insensitive)
//	Regex("description", "mongodb|database") // Contains "mongodb" or "database"
func Regex(field, pattern string, options ...string) Filter {
	opts := ""
	if len(options) > 0 {
		opts = options[0]
	}
	return regexFilter{field: field, pattern: pattern, options: opts}
}

type regexFilter struct {
	field   string
	pattern string
	options string
}

func (f regexFilter) ToMongo() bson.M {
	if f.options == "" {
		return bson.M{f.field: bson.M{"$regex": f.pattern}}
	}
	return bson.M{f.field: bson.M{"$regex": f.pattern, "$options": f.options}}
}

// All creates a filter that matches documents where the array field contains all specified values.
// The order of values doesn't matter, but all values must be present.
//
// MongoDB equivalent: {field: {$all: [values...]}}
//
// Example:
//
//	All("tags", []string{"mongodb", "database", "nosql"})
//	// Matches documents where tags array contains all three values
func All(field string, values any) Filter {
	return opFilter{field: field, op: "$all", value: values}
}

// Size creates a filter that matches documents where the array field has exactly the specified size.
//
// MongoDB equivalent: {field: {$size: size}}
//
// Example:
//
//	Size("items", 3)           // {"items": {"$size": 3}}
//	Size("tags", 0)            // Match documents with empty tags array
func Size(field string, size int) Filter {
	return opFilter{field: field, op: "$size", value: size}
}

// ElemMatch creates a filter that matches documents where at least one array element
// satisfies all the specified filter criteria. This is useful for querying arrays
// of embedded documents.
//
// MongoDB equivalent: {field: {$elemMatch: filter}}
//
// Example:
//
//	// Match orders with at least one item that costs >= $100 and quantity > 5
//	ElemMatch("items", And(
//	    Gte("price", 100),
//	    Gt("quantity", 5),
//	))
//
//	// Match users with a score between 80-90
//	ElemMatch("scores", And(Gte("score", 80), Lt("score", 90)))
func ElemMatch(field string, filter Filter) Filter {
	return elemMatchFilter{field: field, filter: filter}
}

type elemMatchFilter struct {
	field  string
	filter Filter
}

func (f elemMatchFilter) ToMongo() bson.M {
	if f.filter == nil {
		return bson.M{f.field: bson.M{"$elemMatch": bson.M{}}}
	}
	return bson.M{f.field: bson.M{"$elemMatch": f.filter.ToMongo()}}
}

// Between creates a filter that matches documents where field is within an inclusive range.
// This is syntactic sugar for And(Gte(field, min), Lte(field, max)).
//
// MongoDB equivalent: {$and: [{field: {$gte: min}}, {field: {$lte: max}}]}
//
// Example:
//
//	Between("age", 18, 65)     // 18 <= age <= 65
//	Between("price", 10.0, 100.0)
//	Between("date", startDate, endDate)
func Between(field string, min, max any) Filter {
	return And(Gte(field, min), Lte(field, max))
}
