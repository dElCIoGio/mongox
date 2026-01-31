# mongox

[![Go Reference](https://pkg.go.dev/badge/github.com/dElCIoGio/mongox.svg)](https://pkg.go.dev/github.com/dElCIoGio/mongox)
[![Go Report Card](https://goreportcard.com/badge/github.com/dElCIoGio/mongox)](https://goreportcard.com/report/github.com/dElCIoGio/mongox)
[![CI](https://github.com/dElCIoGio/mongox/actions/workflows/ci.yml/badge.svg)](https://github.com/dElCIoGio/mongox/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go-native, specification-driven ODM for MongoDB. Build composable, type-safe queries with the specification pattern.

## Features

- **Generic Repository Pattern** - Type-safe CRUD operations with Go generics
- **Specification Pattern** - Composable, reusable query filters
- **Fluent Aggregation Builder** - Type-safe aggregation pipelines
- **Automatic Timestamps** - `CreatedAt` and `UpdatedAt` managed automatically
- **Lifecycle Hooks** - `BeforeSave` and `AfterLoad` hooks for custom logic
- **Soft Delete** - Built-in soft delete support with automatic filtering
- **Bulk Operations** - Efficient batch inserts, updates, and deletes
- **Transactions** - Full transaction support with automatic retry
- **Pagination** - Built-in pagination with metadata
- **Index Management** - Declarative index definitions
- **Validation** - Document validation before persistence

## Installation

```bash
go get github.com/dElCIoGio/mongox
```

Requires Go 1.21 or later.

## Quick Start

### Define Your Document

```go
package main

import (
    "github.com/dElCIoGio/mongox/document"
)

type User struct {
    document.Base `bson:",inline"`

    Name  string `bson:"name"`
    Email string `bson:"email"`
    Age   int    `bson:"age"`
}
```

### Create a Repository

```go
import (
    "context"

    mongorepo "github.com/dElCIoGio/mongox/repository/mongo"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    ctx := context.Background()

    // Connect to MongoDB
    client, _ := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
    coll := client.Database("myapp").Collection("users")

    // Create repository
    repo := mongorepo.New[User](coll)

    // Insert a document
    user := &User{Name: "John", Email: "john@example.com", Age: 30}
    repo.InsertOne(ctx, user)
    // user.ID, user.CreatedAt, user.UpdatedAt are automatically set
}
```

### Build Queries with Specifications

```go
import "github.com/dElCIoGio/mongox/spec"

// Simple filter
users, _ := repo.Find(ctx, spec.Eq("name", "John"))

// Compose complex filters
filter := spec.And(
    spec.Gte("age", 18),
    spec.Lt("age", 65),
    spec.Or(
        spec.Eq("status", "active"),
        spec.Eq("role", "admin"),
    ),
)
users, _ := repo.Find(ctx, filter)

// Reusable specifications
var (
    IsAdult    = spec.Gte("age", 18)
    IsActive   = spec.Eq("status", "active")
    ActiveAdult = spec.And(IsAdult, IsActive)
)

users, _ := repo.Find(ctx, ActiveAdult)
```

## Core Concepts

### Filter Operators

```go
// Comparison
spec.Eq("field", value)      // field == value
spec.Ne("field", value)      // field != value
spec.Gt("field", value)      // field > value
spec.Gte("field", value)     // field >= value
spec.Lt("field", value)      // field < value
spec.Lte("field", value)     // field <= value

// Array
spec.In("field", []any{...})      // field in [...]
spec.NotIn("field", []any{...})   // field not in [...]
spec.All("tags", []string{...})   // array contains all
spec.Size("tags", 3)              // array has size 3

// Logical
spec.And(filter1, filter2, ...)   // all conditions
spec.Or(filter1, filter2, ...)    // any condition
spec.Not(filter)                  // negate condition

// Pattern
spec.Regex("email", "@gmail\\.com$", "i")  // regex match

// Existence
spec.Exists("field", true)    // field exists

// Range helper
spec.Between("age", 18, 65)   // 18 <= age <= 65

// Array element matching
spec.ElemMatch("results", spec.Gte("score", 80))
```

### Update Operators

```go
// Field updates
spec.Set("status", "active")
spec.SetFields(bson.M{"name": "John", "age": 30})
spec.Unset("temporary_field")
spec.Rename("old_name", "new_name")

// Numeric operations
spec.Inc("counter", 1)
spec.Mul("price", 1.1)
spec.Min("low_score", score)
spec.Max("high_score", score)

// Array operations
spec.Push("tags", "new-tag")
spec.Pull("tags", "old-tag")
spec.AddToSet("tags", "unique-tag")
spec.PopFirst("queue")
spec.PopLast("stack")

// Combine multiple updates
update := spec.Combine(
    spec.Set("status", "processed"),
    spec.Inc("process_count", 1),
    spec.Push("history", "processed"),
)
repo.UpdateOne(ctx, filter, update)
```

### Aggregation Pipelines

```go
pipeline := spec.NewPipeline().
    Match(spec.Eq("status", "active")).
    Group(bson.M{
        "_id":   "$category",
        "total": spec.Sum("$amount"),
        "count": spec.Sum(1),
        "avg":   spec.Avg("$amount"),
    }).
    Sort(bson.D{{"total", -1}}).
    Limit(10)

results, _ := repo.AggregateRaw(ctx, pipeline)
```

### Lifecycle Hooks

```go
type User struct {
    document.Base `bson:",inline"`
    Email string `bson:"email"`
}

// Called before insert/replace
func (u *User) BeforeSave(ctx context.Context) error {
    u.Email = strings.ToLower(u.Email)
    return nil
}

// Called after find operations
func (u *User) AfterLoad(ctx context.Context) error {
    // Compute derived fields
    return nil
}
```

### Validation

```go
type User struct {
    document.Base `bson:",inline"`
    Email string `bson:"email"`
    Age   int    `bson:"age"`
}

func (u *User) Validate() error {
    var errs document.MultiValidationError

    if u.Email == "" {
        errs = append(errs, document.ValidationError{
            Field: "email", Message: "required",
        })
    }

    if u.Age < 0 {
        errs = append(errs, document.ValidationError{
            Field: "age", Message: "must be positive",
        })
    }

    if len(errs) > 0 {
        return errs
    }
    return nil
}
```

### Index Management

```go
type User struct {
    document.Base `bson:",inline"`
    Email string `bson:"email"`
}

func (u User) Indexes() []document.Index {
    return []document.Index{
        {
            Keys:   bson.D{{"email", 1}},
            Unique: true,
            Name:   "unique_email",
        },
    }
}

// Create repository with automatic index creation
repo, err := mongorepo.NewWithIndexes[User](ctx, coll)
```

### Soft Delete

```go
type Post struct {
    document.Base        `bson:",inline"`
    document.SoftDeletable `bson:",inline"`

    Title   string `bson:"title"`
    Content string `bson:"content"`
}

// Create soft delete repository
repo := mongorepo.NewSoftDelete[Post](coll)

// Soft delete (sets deleted_at)
repo.SoftDelete(ctx, spec.Eq("_id", postID))

// Find excludes soft-deleted by default
posts, _ := repo.Find(ctx, nil)

// Include soft-deleted
posts, _ := repo.FindWithDeleted(ctx, nil)

// Restore soft-deleted
repo.Restore(ctx, spec.Eq("_id", postID))

// Permanently delete
repo.Purge(ctx, spec.Lt("deleted_at", cutoffDate))
```

### Transactions

```go
err := mongorepo.RunInTransaction(ctx, client, func(txCtx context.Context) error {
    // All operations use txCtx
    if err := userRepo.InsertOne(txCtx, user); err != nil {
        return err // Transaction rolls back
    }

    if err := orderRepo.InsertOne(txCtx, order); err != nil {
        return err // Transaction rolls back
    }

    return nil // Transaction commits
})
```

### Pagination

```go
page, err := repo.FindPaginated(ctx, filter, 1, 20) // page 1, 20 per page

fmt.Printf("Page %d of %d\n", page.Page, page.TotalPages)
fmt.Printf("Total: %d items\n", page.Total)
fmt.Printf("Has next: %v\n", page.HasNext)

for _, item := range page.Items {
    // Process items
}
```

### Client Management

```go
import "github.com/dElCIoGio/mongox/client"

c, err := client.Connect(ctx, "mongodb://localhost:27017",
    client.WithDatabase("myapp"),
    client.WithMaxPoolSize(100),
    client.WithAppName("my-service"),
)
if err != nil {
    log.Fatal(err)
}
defer c.Close(ctx)

// Get collection
coll := c.Collection("users")
```

## Examples

See the [examples](./examples) directory for complete working examples:

- [Basic CRUD](./examples/basic) - Simple create, read, update, delete operations
- [Specifications](./examples/specifications) - Composing complex filters
- [Hooks](./examples/hooks) - Using lifecycle hooks for validation and transformation
- [Transactions](./examples/transactions) - Atomic operations across documents
- [Aggregation](./examples/aggregation) - Building aggregation pipelines

## API Reference

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/dElCIoGio/mongox).

### Packages

| Package | Description |
|---------|-------------|
| `document` | Base types, hooks, validation, soft delete |
| `spec` | Filter operators, update operators, pipeline builder |
| `repository` | Repository interface and options |
| `repository/mongo` | MongoDB implementation |
| `client` | Connection management |

## Future Improvements

The following features are planned for future releases:

### Observability
- OpenTelemetry tracing integration
- Metrics collection for repository operations
- Structured logging support

### Query Enhancements
- Text search operators
- Geospatial query support
- Change streams integration

### Performance
- Query result caching layer
- Connection pool optimization
- Batch operation improvements

### Developer Experience
- Code generation for type-safe field references
- Migration tooling
- CLI for common operations

### Multi-tenancy
- Built-in tenant scoping
- Automatic tenant filtering
- Tenant-aware indexes

Contributions are welcome! See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## Design Philosophy

- **Explicit over implicit** - No hidden behavior or magic
- **Composition over inheritance** - Build complex queries from simple parts
- **Type safety** - Leverage Go generics for compile-time checks
- **Minimal abstraction** - Stay close to MongoDB's power
- **Testability** - Business rules can be tested without a database

## Acknowledgments

Inspired by:
- [Specification Pattern](https://martinfowler.com/apsupp/spec.pdf) (Domain-Driven Design)
- [Beanie](https://github.com/roman-right/beanie) (Python ODM)
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver)
