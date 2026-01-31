package document

import "context"

// BeforeSave is an optional interface that documents can implement to perform
// validation or transformation before being saved to MongoDB.
//
// The BeforeSave hook is called automatically by the repository before:
//   - InsertOne
//   - ReplaceOne
//
// If BeforeSave returns an error, the operation is aborted and the error is returned.
//
// Common use cases:
//   - Validation (checking required fields, value ranges)
//   - Normalization (trimming strings, formatting data)
//   - Computing derived fields
//   - Setting default values
//
// Example:
//
//	func (u *User) BeforeSave(ctx context.Context) error {
//	    u.Email = strings.ToLower(strings.TrimSpace(u.Email))
//	    if u.Email == "" {
//	        return errors.New("email is required")
//	    }
//	    return nil
//	}
type BeforeSave interface {
	BeforeSave(ctx context.Context) error
}

// AfterLoad is an optional interface that documents can implement to perform
// transformations after being loaded from MongoDB.
//
// The AfterLoad hook is called automatically by the repository after:
//   - FindOne
//   - Find (for each document)
//   - Aggregate (for each document, when using typed results)
//
// If AfterLoad returns an error, the operation returns the error and no documents.
//
// Common use cases:
//   - Decrypting sensitive fields
//   - Computing derived/virtual fields
//   - Resolving lazy-loaded references
//   - Unmarshaling custom data formats
//
// Example:
//
//	func (u *User) AfterLoad(ctx context.Context) error {
//	    u.FullName = u.FirstName + " " + u.LastName
//	    return nil
//	}
type AfterLoad interface {
	AfterLoad(ctx context.Context) error
}
