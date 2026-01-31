package document_test

import (
	"testing"
	"time"

	"github.com/dElCIoGio/mongox/document"
)

func TestSoftDeletable_IsDeleted(t *testing.T) {
	t.Run("not deleted", func(t *testing.T) {
		s := &document.SoftDeletable{}
		if s.IsDeleted() {
			t.Fatal("expected IsDeleted() to be false for new SoftDeletable")
		}
	})

	t.Run("deleted", func(t *testing.T) {
		now := time.Now()
		s := &document.SoftDeletable{DeletedAt: &now}
		if !s.IsDeleted() {
			t.Fatal("expected IsDeleted() to be true when DeletedAt is set")
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var s *document.SoftDeletable
		if s.IsDeleted() {
			t.Fatal("expected IsDeleted() to be false for nil receiver")
		}
	})
}

func TestSoftDeletable_MarkDeleted(t *testing.T) {
	t.Run("with time", func(t *testing.T) {
		s := &document.SoftDeletable{}
		now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

		s.MarkDeleted(now)

		if s.DeletedAt == nil {
			t.Fatal("expected DeletedAt to be set")
		}
		if !s.DeletedAt.Equal(now) {
			t.Fatalf("expected DeletedAt to be %v, got %v", now, *s.DeletedAt)
		}
	})

	t.Run("with zero time", func(t *testing.T) {
		s := &document.SoftDeletable{}
		before := time.Now().UTC()

		s.MarkDeleted(time.Time{})

		after := time.Now().UTC()

		if s.DeletedAt == nil {
			t.Fatal("expected DeletedAt to be set")
		}
		if s.DeletedAt.Before(before) || s.DeletedAt.After(after) {
			t.Fatalf("expected DeletedAt to be between %v and %v, got %v", before, after, *s.DeletedAt)
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var s *document.SoftDeletable
		// Should not panic
		s.MarkDeleted(time.Now())
	})
}

func TestSoftDeletable_Restore(t *testing.T) {
	t.Run("restore deleted", func(t *testing.T) {
		now := time.Now()
		s := &document.SoftDeletable{DeletedAt: &now}

		s.Restore()

		if s.DeletedAt != nil {
			t.Fatal("expected DeletedAt to be nil after Restore()")
		}
		if s.IsDeleted() {
			t.Fatal("expected IsDeleted() to be false after Restore()")
		}
	})

	t.Run("restore non-deleted", func(t *testing.T) {
		s := &document.SoftDeletable{}
		s.Restore()

		if s.DeletedAt != nil {
			t.Fatal("expected DeletedAt to remain nil")
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var s *document.SoftDeletable
		// Should not panic
		s.Restore()
	})
}

func TestSoftDeletable_Workflow(t *testing.T) {
	s := &document.SoftDeletable{}

	// Initially not deleted
	if s.IsDeleted() {
		t.Fatal("should not be deleted initially")
	}

	// Mark as deleted
	s.MarkDeleted(time.Now())
	if !s.IsDeleted() {
		t.Fatal("should be deleted after MarkDeleted()")
	}

	// Restore
	s.Restore()
	if s.IsDeleted() {
		t.Fatal("should not be deleted after Restore()")
	}

	// Delete again
	s.MarkDeleted(time.Now())
	if !s.IsDeleted() {
		t.Fatal("should be deleted after second MarkDeleted()")
	}
}
