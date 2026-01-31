package repository_test

import (
	"testing"

	"github.com/dElCIoGio/mongox/repository"
)

func TestPaginationOptions_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    repository.PaginationOptions
		wantPage int
		wantPer  int
	}{
		{
			name:     "zero values get defaults",
			input:    repository.PaginationOptions{},
			wantPage: 1,
			wantPer:  20,
		},
		{
			name:     "negative page becomes 1",
			input:    repository.PaginationOptions{Page: -5, PerPage: 10},
			wantPage: 1,
			wantPer:  10,
		},
		{
			name:     "perPage exceeds max gets capped",
			input:    repository.PaginationOptions{Page: 1, PerPage: 200, MaxPerPage: 100},
			wantPage: 1,
			wantPer:  100,
		},
		{
			name:     "custom defaults are respected",
			input:    repository.PaginationOptions{DefaultPerPage: 50, MaxPerPage: 200},
			wantPage: 1,
			wantPer:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Normalize()
			if tt.input.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", tt.input.Page, tt.wantPage)
			}
			if tt.input.PerPage != tt.wantPer {
				t.Errorf("PerPage = %d, want %d", tt.input.PerPage, tt.wantPer)
			}
		})
	}
}

func TestPaginationOptions_Skip(t *testing.T) {
	tests := []struct {
		name    string
		page    int
		perPage int
		want    int64
	}{
		{"first page", 1, 20, 0},
		{"second page", 2, 20, 20},
		{"third page", 3, 10, 20},
		{"large page", 100, 50, 4950},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := repository.PaginationOptions{Page: tt.page, PerPage: tt.perPage}
			if got := opts.Skip(); got != tt.want {
				t.Errorf("Skip() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPaginationOptions_Limit(t *testing.T) {
	opts := repository.PaginationOptions{Page: 1, PerPage: 25}
	if got := opts.Limit(); got != 25 {
		t.Errorf("Limit() = %d, want 25", got)
	}
}

func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		name    string
		total   int64
		perPage int
		want    int
	}{
		{"exact division", 100, 20, 5},
		{"with remainder", 101, 20, 6},
		{"less than one page", 5, 20, 1},
		{"zero total", 0, 20, 0},
		{"one item", 1, 20, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repository.CalculateTotalPages(tt.total, tt.perPage); got != tt.want {
				t.Errorf("CalculateTotalPages(%d, %d) = %d, want %d", tt.total, tt.perPage, got, tt.want)
			}
		})
	}
}

func TestPage_Methods(t *testing.T) {
	t.Run("empty page", func(t *testing.T) {
		page := &repository.Page[string]{Items: []string{}}
		if !page.IsEmpty() {
			t.Error("expected IsEmpty() to be true")
		}
	})

	t.Run("non-empty page", func(t *testing.T) {
		page := &repository.Page[string]{Items: []string{"a", "b"}}
		if page.IsEmpty() {
			t.Error("expected IsEmpty() to be false")
		}
	})

	t.Run("first page", func(t *testing.T) {
		page := &repository.Page[string]{Page: 1, TotalPages: 5}
		if !page.FirstPage() {
			t.Error("expected FirstPage() to be true")
		}
		if page.LastPage() {
			t.Error("expected LastPage() to be false")
		}
	})

	t.Run("last page", func(t *testing.T) {
		page := &repository.Page[string]{Page: 5, TotalPages: 5}
		if page.FirstPage() {
			t.Error("expected FirstPage() to be false")
		}
		if !page.LastPage() {
			t.Error("expected LastPage() to be true")
		}
	})

	t.Run("middle page", func(t *testing.T) {
		page := &repository.Page[string]{Page: 3, TotalPages: 5, HasNext: true, HasPrev: true}
		if page.FirstPage() {
			t.Error("expected FirstPage() to be false")
		}
		if page.LastPage() {
			t.Error("expected LastPage() to be false")
		}
		if !page.HasNext {
			t.Error("expected HasNext to be true")
		}
		if !page.HasPrev {
			t.Error("expected HasPrev to be true")
		}
	})
}

func TestDefaultPaginationOptions(t *testing.T) {
	opts := repository.DefaultPaginationOptions()

	if opts.Page != 1 {
		t.Errorf("Page = %d, want 1", opts.Page)
	}
	if opts.PerPage != 20 {
		t.Errorf("PerPage = %d, want 20", opts.PerPage)
	}
	if opts.MaxPerPage != 100 {
		t.Errorf("MaxPerPage = %d, want 100", opts.MaxPerPage)
	}
	if opts.DefaultPerPage != 20 {
		t.Errorf("DefaultPerPage = %d, want 20", opts.DefaultPerPage)
	}
}
