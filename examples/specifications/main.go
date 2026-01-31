// Example: Composing Filters with the Specification Pattern
//
// This example demonstrates how to build complex, composable queries using
// the spec package's filter operators and logical combinators.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dElCIoGio/mongox/document"
	"github.com/dElCIoGio/mongox/repository"
	mongorepo "github.com/dElCIoGio/mongox/repository/mongo"
	"github.com/dElCIoGio/mongox/spec"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Product represents a product in an e-commerce system.
type Product struct {
	document.Base `bson:",inline"`

	Name       string   `bson:"name"`
	Category   string   `bson:"category"`
	Price      float64  `bson:"price"`
	InStock    bool     `bson:"in_stock"`
	Tags       []string `bson:"tags"`
	Rating     float64  `bson:"rating"`
	ReviewCount int     `bson:"review_count"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Setup
	coll := client.Database("example_db").Collection("products")
	repo := mongorepo.New[Product](coll)
	coll.Drop(ctx)

	// Seed data
	seedProducts(ctx, repo)

	// ========== COMPARISON OPERATORS ==========
	fmt.Println("=== COMPARISON OPERATORS ===\n")

	// Eq - Exact match
	electronics, _ := repo.Find(ctx, spec.Eq("category", "electronics"))
	fmt.Printf("Electronics (Eq): %d products\n", len(electronics))

	// Ne - Not equal
	notElectronics, _ := repo.Find(ctx, spec.Ne("category", "electronics"))
	fmt.Printf("Not Electronics (Ne): %d products\n", len(notElectronics))

	// Gt, Gte, Lt, Lte - Comparisons
	expensive, _ := repo.Find(ctx, spec.Gt("price", 500))
	fmt.Printf("Price > $500 (Gt): %d products\n", len(expensive))

	affordable, _ := repo.Find(ctx, spec.Lte("price", 100))
	fmt.Printf("Price <= $100 (Lte): %d products\n", len(affordable))

	// Between - Range query
	midRange, _ := repo.Find(ctx, spec.Between("price", 100, 500))
	fmt.Printf("Price $100-$500 (Between): %d products\n", len(midRange))

	// ========== ARRAY OPERATORS ==========
	fmt.Println("\n=== ARRAY OPERATORS ===\n")

	// In - Match any value in list
	categories := []string{"electronics", "books"}
	inCategories, _ := repo.Find(ctx, spec.In("category", categories))
	fmt.Printf("In electronics or books (In): %d products\n", len(inCategories))

	// NotIn - Match none of the values
	notInCategories, _ := repo.Find(ctx, spec.NotIn("category", categories))
	fmt.Printf("Not in electronics or books (NotIn): %d products\n", len(notInCategories))

	// All - Array contains all values
	withAllTags, _ := repo.Find(ctx, spec.All("tags", []string{"featured", "bestseller"}))
	fmt.Printf("Has both 'featured' AND 'bestseller' tags (All): %d products\n", len(withAllTags))

	// Size - Array has exact size
	twoTags, _ := repo.Find(ctx, spec.Size("tags", 2))
	fmt.Printf("Has exactly 2 tags (Size): %d products\n", len(twoTags))

	// ========== LOGICAL OPERATORS ==========
	fmt.Println("\n=== LOGICAL OPERATORS ===\n")

	// And - All conditions must match
	inStockAndAffordable, _ := repo.Find(ctx, spec.And(
		spec.Eq("in_stock", true),
		spec.Lte("price", 100),
	))
	fmt.Printf("In stock AND price <= $100 (And): %d products\n", len(inStockAndAffordable))

	// Or - Any condition can match
	cheapOrRated, _ := repo.Find(ctx, spec.Or(
		spec.Lt("price", 50),
		spec.Gte("rating", 4.5),
	))
	fmt.Printf("Price < $50 OR rating >= 4.5 (Or): %d products\n", len(cheapOrRated))

	// Not - Negate a condition
	notExpensive, _ := repo.Find(ctx, spec.Not(spec.Gt("price", 500)))
	fmt.Printf("NOT price > $500 (Not): %d products\n", len(notExpensive))

	// ========== COMPLEX COMPOSITIONS ==========
	fmt.Println("\n=== COMPLEX COMPOSITIONS ===\n")

	// Complex filter: (in_stock AND price < 200) OR (rating >= 4.5 AND review_count > 100)
	complexFilter := spec.Or(
		spec.And(
			spec.Eq("in_stock", true),
			spec.Lt("price", 200),
		),
		spec.And(
			spec.Gte("rating", 4.5),
			spec.Gt("review_count", 100),
		),
	)
	complexResults, _ := repo.Find(ctx, complexFilter)
	fmt.Printf("Complex query: %d products\n", len(complexResults))

	// ========== REGEX PATTERN MATCHING ==========
	fmt.Println("\n=== REGEX MATCHING ===\n")

	// Case-insensitive search
	proProducts, _ := repo.Find(ctx, spec.Regex("name", "pro", "i"))
	fmt.Printf("Name contains 'pro' (case-insensitive): %d products\n", len(proProducts))

	// Starts with pattern
	startsWithS, _ := repo.Find(ctx, spec.Regex("name", "^S"))
	fmt.Printf("Name starts with 'S': %d products\n", len(startsWithS))

	// ========== EXISTS OPERATOR ==========
	fmt.Println("\n=== EXISTS OPERATOR ===\n")

	hasRating, _ := repo.Find(ctx, spec.Exists("rating", true))
	fmt.Printf("Has rating field: %d products\n", len(hasRating))

	// ========== REUSABLE SPECIFICATIONS ==========
	fmt.Println("\n=== REUSABLE SPECIFICATIONS ===\n")

	// Define reusable filters
	inStockFilter := spec.Eq("in_stock", true)
	highlyRatedFilter := spec.Gte("rating", 4.0)
	affordableFilter := spec.Lte("price", 200)

	// Compose them as needed
	bestDeals, _ := repo.Find(ctx, spec.And(inStockFilter, highlyRatedFilter, affordableFilter),
		repository.WithSort(bson.D{{"rating", -1}}),
		repository.WithLimit(5),
	)
	fmt.Println("Top 5 best deals (in stock, highly rated, affordable):")
	for _, p := range bestDeals {
		fmt.Printf("  - %s: $%.2f (%.1f stars)\n", p.Name, p.Price, p.Rating)
	}
}

func seedProducts(ctx context.Context, repo *mongorepo.MongoRepository[Product]) {
	products := []*Product{
		{Name: "Smartphone Pro", Category: "electronics", Price: 799.99, InStock: true, Tags: []string{"featured", "bestseller"}, Rating: 4.5, ReviewCount: 250},
		{Name: "Laptop Elite", Category: "electronics", Price: 1299.99, InStock: true, Tags: []string{"featured"}, Rating: 4.8, ReviewCount: 180},
		{Name: "Wireless Earbuds", Category: "electronics", Price: 79.99, InStock: true, Tags: []string{"bestseller"}, Rating: 4.2, ReviewCount: 500},
		{Name: "Smart Watch", Category: "electronics", Price: 299.99, InStock: false, Tags: []string{"featured", "bestseller"}, Rating: 4.6, ReviewCount: 320},
		{Name: "Go Programming", Category: "books", Price: 45.99, InStock: true, Tags: []string{"educational"}, Rating: 4.9, ReviewCount: 150},
		{Name: "System Design", Category: "books", Price: 55.99, InStock: true, Tags: []string{"educational", "bestseller"}, Rating: 4.7, ReviewCount: 200},
		{Name: "Running Shoes", Category: "sports", Price: 89.99, InStock: true, Tags: []string{"featured"}, Rating: 4.3, ReviewCount: 400},
		{Name: "Yoga Mat", Category: "sports", Price: 29.99, InStock: true, Tags: []string{}, Rating: 4.1, ReviewCount: 80},
		{Name: "Coffee Maker Pro", Category: "home", Price: 149.99, InStock: true, Tags: []string{"bestseller"}, Rating: 4.4, ReviewCount: 220},
		{Name: "Desk Lamp", Category: "home", Price: 34.99, InStock: false, Tags: []string{}, Rating: 3.9, ReviewCount: 45},
	}

	repo.InsertMany(ctx, products)
	fmt.Printf("Seeded %d products\n\n", len(products))
}
