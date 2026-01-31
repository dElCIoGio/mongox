// Package spec provides composable MongoDB query filters using the specification pattern.
//
// The package focuses on type-safe, composable filters that translate directly to MongoDB
// queries. Use the filter operators (Eq, Gt, In, etc.) and logical combinators (And, Or, Not)
// to build complex queries in a readable, maintainable way.
//
// Example:
//
//	filter := spec.And(
//	    spec.Eq("status", "active"),
//	    spec.Gte("age", 18),
//	    spec.Or(
//	        spec.Eq("role", "admin"),
//	        spec.Eq("role", "moderator"),
//	    ),
//	)
//	users, err := repo.Find(ctx, filter, nil)
package spec
