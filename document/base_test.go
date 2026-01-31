package document_test

import (
	"testing"
	"time"

	"github.com/dElCIoGio/mongox/document"
)

func TestTouchForInsert(t *testing.T) {
	var b document.Base
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	b.TouchForInsert(now)

	if b.ID.IsZero() {
		t.Fatal("expected ID to be set")
	}
	if b.CreatedAt.IsZero() || !b.CreatedAt.Equal(now) {
		t.Fatal("expected CreatedAt to be set to now")
	}
	if b.UpdatedAt.IsZero() || !b.UpdatedAt.Equal(now) {
		t.Fatal("expected UpdatedAt to be set to now")
	}
}

func TestTouchForUpdate(t *testing.T) {
	now1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	b := document.Base{}
	b.TouchForInsert(now1)
	b.TouchForUpdate(now2)

	if !b.CreatedAt.Equal(now1) {
		t.Fatal("expected CreatedAt to remain unchanged")
	}
	if !b.UpdatedAt.Equal(now2) {
		t.Fatal("expected UpdatedAt to be updated")
	}
}
