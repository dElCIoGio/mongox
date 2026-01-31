package document

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// Index defines a MongoDB index specification.
// Use this struct to declare indexes that should be created on a collection.
//
// Example:
//
//	func (u User) Indexes() []document.Index {
//	    return []document.Index{
//	        {Keys: bson.D{{"email", 1}}, Unique: true, Name: "unique_email"},
//	        {Keys: bson.D{{"created_at", -1}}, Name: "created_at_desc"},
//	    }
//	}
type Index struct {
	// Keys defines the index keys and their sort order.
	// Use 1 for ascending, -1 for descending.
	// For compound indexes, add multiple key-value pairs.
	//
	// Examples:
	//   bson.D{{"email", 1}}                    // Single field ascending
	//   bson.D{{"status", 1}, {"created_at", -1}} // Compound index
	//   bson.D{{"location", "2dsphere"}}        // Geospatial index
	Keys bson.D

	// Unique enforces uniqueness constraint on the indexed fields.
	// Attempting to insert a duplicate value will return an error.
	Unique bool

	// Sparse creates a sparse index that only includes documents
	// that contain the indexed field(s), even if the field value is null.
	// Useful for optional fields with unique constraints.
	Sparse bool

	// Name specifies a custom name for the index.
	// If not provided, MongoDB generates a name based on the keys.
	Name string

	// TTL sets the time-to-live for documents in the collection.
	// Documents are automatically deleted after TTL duration from
	// the indexed date field. Only works on single-field indexes
	// with date type values.
	//
	// Example:
	//   Index{Keys: bson.D{{"expires_at", 1}}, TTL: 24 * time.Hour}
	TTL *time.Duration

	// Background specifies whether the index should be built in the background.
	// Note: In MongoDB 4.2+, all index builds are background by default.
	Background bool

	// PartialFilterExpression creates a partial index that only indexes
	// documents matching the filter expression.
	//
	// Example:
	//   Index{
	//       Keys: bson.D{{"email", 1}},
	//       PartialFilterExpression: bson.M{"status": "active"},
	//   }
	PartialFilterExpression bson.M
}

// Indexed is an interface that documents can implement to declare
// indexes that should exist on their collection.
//
// When using NewIndexedRepository, the indexes returned by this method
// are automatically created on the MongoDB collection.
//
// Example:
//
//	type User struct {
//	    document.Base `bson:",inline"`
//	    Email    string `bson:"email"`
//	    Username string `bson:"username"`
//	    Status   string `bson:"status"`
//	}
//
//	func (u User) Indexes() []document.Index {
//	    return []document.Index{
//	        {
//	            Keys:   bson.D{{"email", 1}},
//	            Unique: true,
//	            Name:   "unique_email",
//	        },
//	        {
//	            Keys:   bson.D{{"username", 1}},
//	            Unique: true,
//	            Sparse: true,
//	            Name:   "unique_username",
//	        },
//	        {
//	            Keys: bson.D{{"status", 1}, {"created_at", -1}},
//	            Name: "status_created",
//	        },
//	    }
//	}
type Indexed interface {
	// Indexes returns the list of indexes that should exist on this
	// document's collection. Called during repository initialization.
	Indexes() []Index
}

// TextIndex creates a text search index specification.
// Text indexes support text search queries on string content.
//
// Example:
//
//	TextIndex("title", "description")
//	// Creates: bson.D{{"title", "text"}, {"description", "text"}}
func TextIndex(fields ...string) bson.D {
	keys := make(bson.D, len(fields))
	for i, field := range fields {
		keys[i] = bson.E{Key: field, Value: "text"}
	}
	return keys
}

// CompoundIndex creates a compound index specification with ascending order.
//
// Example:
//
//	CompoundIndex("status", "created_at")
//	// Creates: bson.D{{"status", 1}, {"created_at", 1}}
func CompoundIndex(fields ...string) bson.D {
	keys := make(bson.D, len(fields))
	for i, field := range fields {
		keys[i] = bson.E{Key: field, Value: 1}
	}
	return keys
}

// DescendingIndex creates a single-field descending index specification.
//
// Example:
//
//	DescendingIndex("created_at")
//	// Creates: bson.D{{"created_at", -1}}
func DescendingIndex(field string) bson.D {
	return bson.D{{Key: field, Value: -1}}
}

// AscendingIndex creates a single-field ascending index specification.
//
// Example:
//
//	AscendingIndex("email")
//	// Creates: bson.D{{"email", 1}}
func AscendingIndex(field string) bson.D {
	return bson.D{{Key: field, Value: 1}}
}

// GeoIndex creates a 2dsphere geospatial index specification.
//
// Example:
//
//	GeoIndex("location")
//	// Creates: bson.D{{"location", "2dsphere"}}
func GeoIndex(field string) bson.D {
	return bson.D{{Key: field, Value: "2dsphere"}}
}
