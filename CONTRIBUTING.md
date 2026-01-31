# Contributing to mongox

Thank you for your interest in contributing to mongox! This document provides guidelines and information for contributors.

## Development Setup

### Prerequisites

- Go 1.21 or later
- MongoDB 4.4 or later (for running tests)
- Git

### Getting Started

1. Fork the repository on GitHub

2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/mongox.git
   cd mongox
   ```

3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/dElCIoGio/mongox.git
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

5. Start a local MongoDB instance:
   ```bash
   # Using Docker
   docker run -d -p 27017:27017 --name mongodb mongo:7.0

   # Or for transactions testing (requires replica set)
   docker run -d -p 27017:27017 --name mongodb mongo:7.0 --replSet rs0
   docker exec mongodb mongosh --eval "rs.initiate()"
   ```

6. Run tests to verify setup:
   ```bash
   go test ./...
   ```

## Code Style Guidelines

### General Principles

- Write idiomatic Go code following [Effective Go](https://go.dev/doc/effective_go)
- Keep functions small and focused on a single responsibility
- Prefer composition over inheritance
- Use meaningful variable and function names

### Formatting

- Run `gofmt` on all code (most editors do this automatically)
- Run `go vet` to catch common mistakes
- Consider using `golangci-lint` for comprehensive linting:
  ```bash
  golangci-lint run
  ```

### Documentation

- All exported types, functions, and methods must have Godoc comments
- Start comments with the name of the element being documented
- Include usage examples in documentation where helpful
- Keep comments concise but informative

```go
// Good
// Filter represents a MongoDB query filter specification.
// Implementations convert to bson.M for use with MongoDB operations.
type Filter interface {
    ToMongo() bson.M
}

// Bad
// This is a filter interface
type Filter interface {
    ToMongo() bson.M
}
```

### Error Handling

- Use the error types defined in `repository/errors.go`
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Check errors immediately after the call that might return them

### Naming Conventions

- Use `MixedCaps` or `mixedCaps` (not underscores)
- Acronyms should be all caps: `ID`, `HTTP`, `URL`
- Interface names should describe behavior, often ending in `-er`
- Avoid stuttering: `spec.Filter` not `spec.SpecFilter`

## Project Structure

```
mongox/
├── document/           # Base document types and hooks
│   ├── base.go         # Base struct with ID and timestamps
│   ├── hooks.go        # BeforeSave/AfterLoad interfaces
│   └── soft_delete.go  # Soft delete functionality
├── repository/         # Repository interfaces and implementations
│   ├── repository.go   # Main Repository interface
│   ├── errors.go       # Error types
│   ├── options.go      # Query options
│   ├── pagination.go   # Pagination helpers
│   ├── bulk.go         # Bulk operation types
│   ├── transaction.go  # Transaction interfaces
│   └── mongo/          # MongoDB implementation
├── spec/               # Specification pattern implementations
│   ├── operators.go    # Filter operators (Eq, Gt, In, etc.)
│   ├── logical.go      # Logical combinators (And, Or, Not)
│   ├── update.go       # Update operators (Set, Inc, Push, etc.)
│   └── pipeline.go     # Aggregation pipeline builder
└── examples/           # Usage examples
```

## Testing Requirements

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./spec/...

# Run a specific test
go test -v -run TestFilterEq ./spec/...

# Run tests with race detector
go test -race ./...

# Run benchmarks
go test -bench=. ./spec/...

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./spec/...
```

### Writing Tests

- Place tests in `*_test.go` files in the same package
- Use table-driven tests for multiple test cases
- Test both success and error cases
- Use descriptive test names

```go
func TestFilterEq(t *testing.T) {
    tests := []struct {
        name     string
        field    string
        value    any
        expected bson.M
    }{
        {
            name:     "string value",
            field:    "status",
            value:    "active",
            expected: bson.M{"status": "active"},
        },
        {
            name:     "integer value",
            field:    "age",
            value:    25,
            expected: bson.M{"age": 25},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            filter := spec.Eq(tt.field, tt.value)
            result := filter.ToMongo()

            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Test Coverage

- Aim for high test coverage on core functionality
- Check coverage with:
  ```bash
  go test -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out
  ```

### Integration Tests

Integration tests that require MongoDB should:
- Use build tags if needed: `//go:build integration`
- Clean up test data after tests complete
- Use unique collection names to avoid conflicts

## Pull Request Process

### Before Submitting

1. **Sync with upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes** following the code style guidelines

4. **Run tests**:
   ```bash
   go test ./...
   ```

5. **Run linters**:
   ```bash
   go vet ./...
   gofmt -s -w .
   ```

6. **Commit your changes**:
   - Write clear, descriptive commit messages
   - Use present tense: "Add feature" not "Added feature"
   - Reference issues: "Fix #123: Handle nil filter case"

### Submitting the PR

1. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Open a Pull Request on GitHub

3. In your PR description:
   - Describe what the PR does
   - Link to any related issues
   - Note any breaking changes
   - Include examples if adding new features

### PR Review Process

- All PRs require at least one approval before merging
- CI must pass (tests, linting)
- Address review feedback promptly
- Keep PRs focused and reasonably sized

### After Merge

- Delete your feature branch
- Sync your fork with upstream

## Reporting Issues

When reporting issues, please include:

- Go version (`go version`)
- MongoDB version
- Operating system
- Minimal code example that reproduces the issue
- Expected vs actual behavior
- Any relevant error messages

## Feature Requests

Feature requests are welcome! Please:

- Check existing issues to avoid duplicates
- Describe the use case and motivation
- Provide examples of the desired API if applicable

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## Questions?

If you have questions about contributing, feel free to:

- Open a GitHub issue with the "question" label
- Check existing issues and discussions

Thank you for contributing to mongox!
