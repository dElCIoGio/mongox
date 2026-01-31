package document

import "time"

// SoftDeletable can be embedded in documents to enable soft delete functionality.
// When soft deleted, documents are marked with a DeletedAt timestamp instead of being removed.
//
// Example:
//
//	type User struct {
//	    document.Base         `bson:",inline"`
//	    document.SoftDeletable `bson:",inline"`
//	    Name  string `bson:"name"`
//	    Email string `bson:"email"`
//	}
type SoftDeletable struct {
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// IsDeleted returns true if the document has been soft deleted.
func (s *SoftDeletable) IsDeleted() bool {
	return s != nil && s.DeletedAt != nil
}

// MarkDeleted sets the DeletedAt timestamp to mark the document as deleted.
func (s *SoftDeletable) MarkDeleted(now time.Time) {
	if s == nil {
		return
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	s.DeletedAt = &now
}

// Restore clears the DeletedAt timestamp to restore the document.
func (s *SoftDeletable) Restore() {
	if s == nil {
		return
	}
	s.DeletedAt = nil
}

// SoftDeletableDoc is an interface for documents that support soft delete.
type SoftDeletableDoc interface {
	IsDeleted() bool
	MarkDeleted(now time.Time)
	Restore()
}
