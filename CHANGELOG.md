# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2024-XX-XX

### Added

#### Core Features
- Generic repository pattern with full CRUD operations (`InsertOne`, `FindOne`, `Find`, `UpdateOne`, `DeleteOne`, `ReplaceOne`)
- Specification pattern for composable, type-safe MongoDB queries
- Automatic timestamp management (`CreatedAt`, `UpdatedAt` via `document.Base`)
- Lifecycle hooks (`BeforeSave`, `AfterLoad`) for custom logic

#### Filter Operators
- Comparison: `Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`
- Array: `In`, `NotIn`, `All`, `Size`, `ElemMatch`
- Logical: `And`, `Or`, `Not` with smart flattening
- Pattern: `Regex` with options support
- Existence: `Exists`
- Range: `Between` helper

#### Update Operators
- Field updates: `Set`, `SetFields`, `Unset`, `Rename`
- Numeric: `Inc`, `Mul`, `Min`, `Max`
- Array: `Push`, `Pull`, `AddToSet`, `PopFirst`, `PopLast`
- Combinator: `Combine` for merging multiple updates

#### Aggregation Pipeline
- Fluent pipeline builder with `NewPipeline()`
- Stages: `Match`, `Project`, `Group`, `GroupBy`, `Sort`, `Limit`, `Skip`, `Unwind`, `Lookup`, `AddFields`, `Count`, `Facet`, `Sample`, `Out`, `Merge`
- Accumulators: `Sum`, `Avg`, `Min`, `Max`, `First`, `Last`, `Push`, `AddToSet`
- Raw pipeline support via `AggregateRaw`

#### Bulk Operations
- `InsertMany` for batch inserts
- `UpdateMany` for batch updates
- `DeleteMany` for batch deletes
- `BulkWrite` for mixed operations

#### Transactions
- `TransactionManager` interface for transaction handling
- `WithTransaction` method with automatic retry
- `RunInTransaction` convenience function

#### Soft Delete
- `SoftDeletable` struct with `DeletedAt` field
- `SoftDeleteRepository` with filtered queries
- Methods: `SoftDelete`, `Restore`, `FindWithDeleted`, `Purge`

#### Pagination
- `FindPaginated` with `Page[T]` result type
- Pagination metadata: `Total`, `TotalPages`, `HasNext`, `HasPrev`

#### Indexes
- `Index` struct for index specification
- `Indexed` interface for document-declared indexes
- `NewWithIndexes` constructor for automatic index creation
- Helper functions: `TextIndex`, `CompoundIndex`, `GeoIndex`

#### Validation
- `Validatable` interface for custom validation
- `ValidationError` and `MultiValidationError` types
- Automatic validation before insert/replace

#### Client Management
- `client.Client` wrapper with connection pooling
- Configuration options: `WithMaxPoolSize`, `WithConnectTimeout`, etc.

#### Error Handling
- Sentinel errors: `ErrNotFound`, `ErrDuplicateKey`, `ErrInvalidFilter`, `ErrValidation`, `ErrNilDocument`, `ErrNilUpdate`

#### Documentation
- Comprehensive Godoc comments on all exported types
- Usage examples in `examples/` directory
- Benchmark tests for performance validation

### Changed
- N/A (initial release)

### Deprecated
- N/A (initial release)

### Removed
- N/A (initial release)

### Fixed
- N/A (initial release)

### Security
- N/A (initial release)

[Unreleased]: https://github.com/dElCIoGio/mongox/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/dElCIoGio/mongox/releases/tag/v0.1.0
