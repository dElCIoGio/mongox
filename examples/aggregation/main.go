// Example: Aggregation Pipelines
//
// This example demonstrates how to use MongoDB aggregation pipelines
// with the type-safe pipeline builder.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dElCIoGio/mongox/document"
	mongorepo "github.com/dElCIoGio/mongox/repository/mongo"
	"github.com/dElCIoGio/mongox/spec"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Sale represents a sales record.
type Sale struct {
	document.Base `bson:",inline"`

	Product   string    `bson:"product"`
	Category  string    `bson:"category"`
	Quantity  int       `bson:"quantity"`
	UnitPrice float64   `bson:"unit_price"`
	Total     float64   `bson:"total"`
	Region    string    `bson:"region"`
	SaleDate  time.Time `bson:"sale_date"`
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
	coll := client.Database("example_db").Collection("sales")
	repo := mongorepo.New[Sale](coll)
	coll.Drop(ctx)

	// Seed data
	seedSales(ctx, repo)

	// ========== BASIC AGGREGATION ==========
	fmt.Println("=== BASIC AGGREGATION ===\n")

	// Total sales by category
	pipeline := spec.NewPipeline().
		Group(bson.M{
			"_id":        "$category",
			"totalSales": spec.Sum("$total"),
			"count":      spec.Sum(1),
			"avgTotal":   spec.Avg("$total"),
		}).
		SortBy("totalSales", -1)

	results, err := repo.AggregateRaw(ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Sales by Category:")
	for _, r := range results {
		fmt.Printf("  %s: $%.2f (%d sales, avg $%.2f)\n",
			r["_id"], r["totalSales"], r["count"], r["avgTotal"])
	}

	// ========== MATCH + GROUP ==========
	fmt.Println("\n=== FILTERED AGGREGATION ===\n")

	// Sales in North region only, grouped by product
	pipeline = spec.NewPipeline().
		Match(spec.Eq("region", "North")).
		GroupBy("$product", bson.M{
			"totalQuantity": spec.Sum("$quantity"),
			"totalRevenue":  spec.Sum("$total"),
		}).
		SortBy("totalRevenue", -1).
		Limit(5)

	results, _ = repo.AggregateRaw(ctx, pipeline)

	fmt.Println("Top 5 Products in North Region:")
	for _, r := range results {
		fmt.Printf("  %s: %d units, $%.2f revenue\n",
			r["_id"], r["totalQuantity"], r["totalRevenue"])
	}

	// ========== DATE AGGREGATION ==========
	fmt.Println("\n=== DATE-BASED AGGREGATION ===\n")

	// Sales by month
	pipeline = spec.NewPipeline().
		Group(bson.M{
			"_id": bson.M{
				"year":  bson.M{"$year": "$sale_date"},
				"month": bson.M{"$month": "$sale_date"},
			},
			"totalSales": spec.Sum("$total"),
			"orderCount": spec.Sum(1),
		}).
		Sort(bson.D{{"_id.year", 1}, {"_id.month", 1}})

	results, _ = repo.AggregateRaw(ctx, pipeline)

	fmt.Println("Monthly Sales:")
	for _, r := range results {
		id := r["_id"].(bson.M)
		fmt.Printf("  %d-%02d: $%.2f (%d orders)\n",
			id["year"], id["month"], r["totalSales"], r["orderCount"])
	}

	// ========== MULTI-STAGE PIPELINE ==========
	fmt.Println("\n=== MULTI-STAGE PIPELINE ===\n")

	// Average order value by region, only for categories with > $500 total
	pipeline = spec.NewPipeline().
		Group(bson.M{
			"_id": bson.M{
				"region":   "$region",
				"category": "$category",
			},
			"totalSales": spec.Sum("$total"),
			"orderCount": spec.Sum(1),
		}).
		MatchRaw(bson.M{"totalSales": bson.M{"$gt": 500}}).
		AddFields(bson.M{
			"avgOrderValue": bson.M{"$divide": []string{"$totalSales", "$orderCount"}},
		}).
		SortBy("avgOrderValue", -1)

	results, _ = repo.AggregateRaw(ctx, pipeline)

	fmt.Println("Regions/Categories with >$500 sales (by avg order value):")
	for _, r := range results {
		id := r["_id"].(bson.M)
		fmt.Printf("  %s/%s: avg $%.2f (total $%.2f, %d orders)\n",
			id["region"], id["category"], r["avgOrderValue"], r["totalSales"], r["orderCount"])
	}

	// ========== UNWIND EXAMPLE ==========
	fmt.Println("\n=== PROJECTION & COMPUTATION ===\n")

	// Calculate profit margin (assuming 30% cost)
	pipeline = spec.NewPipeline().
		Project(bson.M{
			"product":  1,
			"category": 1,
			"total":    1,
			"cost":     bson.M{"$multiply": []any{"$total", 0.7}},
			"profit":   bson.M{"$multiply": []any{"$total", 0.3}},
		}).
		Group(bson.M{
			"_id":         "$category",
			"totalProfit": spec.Sum("$profit"),
			"totalCost":   spec.Sum("$cost"),
		}).
		SortBy("totalProfit", -1)

	results, _ = repo.AggregateRaw(ctx, pipeline)

	fmt.Println("Profit by Category:")
	for _, r := range results {
		fmt.Printf("  %s: profit $%.2f (cost $%.2f)\n",
			r["_id"], r["totalProfit"], r["totalCost"])
	}

	// ========== USING RAW BSON PIPELINE ==========
	fmt.Println("\n=== RAW BSON PIPELINE ===\n")

	// You can also use raw bson.M slices for complex pipelines
	rawPipeline := []bson.M{
		{"$match": bson.M{"quantity": bson.M{"$gte": 3}}},
		{"$group": bson.M{
			"_id":       nil,
			"totalSales": bson.M{"$sum": "$total"},
			"avgQuantity": bson.M{"$avg": "$quantity"},
		}},
	}

	results, _ = repo.AggregateRaw(ctx, rawPipeline)

	if len(results) > 0 {
		fmt.Printf("High-quantity orders (qty >= 3):\n")
		fmt.Printf("  Total: $%.2f, Avg Quantity: %.1f\n",
			results[0]["totalSales"], results[0]["avgQuantity"])
	}
}

func seedSales(ctx context.Context, repo *mongorepo.MongoRepository[Sale]) {
	now := time.Now()
	sales := []*Sale{
		// Electronics - North
		{Product: "Laptop", Category: "Electronics", Quantity: 2, UnitPrice: 999.99, Total: 1999.98, Region: "North", SaleDate: now.AddDate(0, -1, 0)},
		{Product: "Phone", Category: "Electronics", Quantity: 5, UnitPrice: 699.99, Total: 3499.95, Region: "North", SaleDate: now.AddDate(0, -1, 5)},
		{Product: "Tablet", Category: "Electronics", Quantity: 3, UnitPrice: 499.99, Total: 1499.97, Region: "North", SaleDate: now.AddDate(0, 0, -10)},

		// Electronics - South
		{Product: "Laptop", Category: "Electronics", Quantity: 1, UnitPrice: 999.99, Total: 999.99, Region: "South", SaleDate: now.AddDate(0, -1, 10)},
		{Product: "Phone", Category: "Electronics", Quantity: 4, UnitPrice: 699.99, Total: 2799.96, Region: "South", SaleDate: now.AddDate(0, 0, -5)},

		// Clothing - North
		{Product: "Jacket", Category: "Clothing", Quantity: 10, UnitPrice: 89.99, Total: 899.90, Region: "North", SaleDate: now.AddDate(0, -1, 3)},
		{Product: "Shoes", Category: "Clothing", Quantity: 8, UnitPrice: 129.99, Total: 1039.92, Region: "North", SaleDate: now.AddDate(0, 0, -2)},

		// Clothing - South
		{Product: "Jacket", Category: "Clothing", Quantity: 6, UnitPrice: 89.99, Total: 539.94, Region: "South", SaleDate: now.AddDate(0, -1, 15)},
		{Product: "Jeans", Category: "Clothing", Quantity: 12, UnitPrice: 59.99, Total: 719.88, Region: "South", SaleDate: now.AddDate(0, 0, -8)},

		// Books - Both regions
		{Product: "Go Programming", Category: "Books", Quantity: 15, UnitPrice: 45.99, Total: 689.85, Region: "North", SaleDate: now.AddDate(0, -2, 0)},
		{Product: "System Design", Category: "Books", Quantity: 8, UnitPrice: 55.99, Total: 447.92, Region: "South", SaleDate: now.AddDate(0, -1, 20)},
	}

	repo.InsertMany(ctx, sales)
	fmt.Printf("Seeded %d sales records\n\n", len(sales))
}
