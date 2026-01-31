package repository

// Page represents a paginated result set.
type Page[T any] struct {
	// Items contains the documents for the current page.
	Items []T

	// Total is the total number of documents matching the filter.
	Total int64

	// Page is the current page number (1-indexed).
	Page int

	// PerPage is the number of items per page.
	PerPage int

	// TotalPages is the total number of pages.
	TotalPages int

	// HasNext indicates if there is a next page.
	HasNext bool

	// HasPrev indicates if there is a previous page.
	HasPrev bool
}

// IsEmpty returns true if the page contains no items.
func (p *Page[T]) IsEmpty() bool {
	return len(p.Items) == 0
}

// FirstPage returns true if this is the first page.
func (p *Page[T]) FirstPage() bool {
	return p.Page == 1
}

// LastPage returns true if this is the last page.
func (p *Page[T]) LastPage() bool {
	return p.Page >= p.TotalPages
}

// PaginationOptions configures pagination behavior.
type PaginationOptions struct {
	// Page is the page number to retrieve (1-indexed). Default is 1.
	Page int

	// PerPage is the number of items per page. Default is 20.
	PerPage int

	// MaxPerPage is the maximum allowed items per page. Default is 100.
	MaxPerPage int

	// DefaultPerPage is used when PerPage is 0 or negative. Default is 20.
	DefaultPerPage int
}

// DefaultPaginationOptions returns default pagination options.
func DefaultPaginationOptions() PaginationOptions {
	return PaginationOptions{
		Page:           1,
		PerPage:        20,
		MaxPerPage:     100,
		DefaultPerPage: 20,
	}
}

// Normalize ensures pagination options are within valid bounds.
func (opts *PaginationOptions) Normalize() {
	if opts.DefaultPerPage <= 0 {
		opts.DefaultPerPage = 20
	}
	if opts.MaxPerPage <= 0 {
		opts.MaxPerPage = 100
	}
	if opts.PerPage <= 0 {
		opts.PerPage = opts.DefaultPerPage
	}
	if opts.PerPage > opts.MaxPerPage {
		opts.PerPage = opts.MaxPerPage
	}
	if opts.Page < 1 {
		opts.Page = 1
	}
}

// Skip returns the number of documents to skip for the current page.
func (opts *PaginationOptions) Skip() int64 {
	opts.Normalize()
	return int64((opts.Page - 1) * opts.PerPage)
}

// Limit returns the number of documents to fetch for the current page.
func (opts *PaginationOptions) Limit() int64 {
	opts.Normalize()
	return int64(opts.PerPage)
}

// CalculateTotalPages calculates the total number of pages for a given total count.
func CalculateTotalPages(total int64, perPage int) int {
	if perPage <= 0 {
		perPage = 20
	}
	pages := int(total) / perPage
	if int(total)%perPage > 0 {
		pages++
	}
	return pages
}
