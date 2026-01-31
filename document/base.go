// Package document provides base types and interfaces for MongoDB documents.
//
// The package includes:
//   - Base: An embeddable struct with ID and timestamp fields
//   - SoftDeletable: An embeddable struct for soft delete functionality
//   - BeforeSave/AfterLoad: Lifecycle hook interfaces
//
// Example usage:
//
//	type User struct {
//	    document.Base         `bson:",inline"`
//	    document.SoftDeletable `bson:",inline"`
//	    Name  string `bson:"name"`
//	    Email string `bson:"email"`
//	}
package document

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Base provides common fields for MongoDB documents: ID, CreatedAt, and UpdatedAt.
// Embed it in your document structs using the inline BSON tag.
//
// When used with MongoRepository:
//   - ID is automatically generated on insert if not set
//   - CreatedAt is automatically set on insert
//   - UpdatedAt is automatically set on insert and update
//
// Example:
//
//	type Order struct {
//	    document.Base `bson:",inline"`
//	    CustomerID string  `bson:"customer_id"`
//	    Total      float64 `bson:"total"`
//	    Status     string  `bson:"status"`
//	}
//
//	// Create and insert
//	order := &Order{CustomerID: "123", Total: 99.99, Status: "pending"}
//	repo.InsertOne(ctx, order)
//	// order.ID, order.CreatedAt, and order.UpdatedAt are now set
type Base struct {
	// ID is the MongoDB document identifier.
	// Automatically generated on insert if not set.
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// CreatedAt records when the document was first created.
	// Automatically set on insert.
	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at"`

	// UpdatedAt records when the document was last modified.
	// Automatically updated on insert and update operations.
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at"`
}

// TouchForInsert sets CreatedAt and UpdatedAt to the given time (or now if zero),
// and generates a new ObjectID if the ID is not set.
//
// This method is called automatically by MongoRepository.InsertOne().
// You typically don't need to call it manually.
func (b *Base) TouchForInsert(now time.Time) {
	if b == nil {
		return
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if b.ID.IsZero() {
		b.ID = primitive.NewObjectID()
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	b.UpdatedAt = now
}

// TouchForUpdate sets UpdatedAt to the given time (or now if zero).
// CreatedAt and ID are preserved.
//
// This method is called automatically by MongoRepository.ReplaceOne().
// You typically don't need to call it manually.
func (b *Base) TouchForUpdate(now time.Time) {
	if b == nil {
		return
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	b.UpdatedAt = now
}
