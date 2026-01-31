// Example: Basic CRUD Operations
//
// This example demonstrates basic Create, Read, Update, and Delete operations
// using the mongox repository pattern.
//
// To run this example, you need a MongoDB instance running locally or update
// the connection string accordingly.
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

// User represents a user document in MongoDB.
// It embeds document.Base for automatic ID and timestamp management.
type User struct {
	document.Base `bson:",inline"`

	Name   string `bson:"name"`
	Email  string `bson:"email"`
	Age    int    `bson:"age"`
	Active bool   `bson:"active"`
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

	// Get collection and create repository
	coll := client.Database("example_db").Collection("users")
	repo := mongorepo.New[User](coll)

	// Clear collection for demo
	coll.Drop(ctx)

	// ========== CREATE ==========
	fmt.Println("=== CREATE ===")

	user := &User{
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    30,
		Active: true,
	}

	if err := repo.InsertOne(ctx, user); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created user: ID=%s, CreatedAt=%v\n", user.ID.Hex(), user.CreatedAt)

	// Insert multiple users
	users := []*User{
		{Name: "Jane Smith", Email: "jane@example.com", Age: 25, Active: true},
		{Name: "Bob Wilson", Email: "bob@example.com", Age: 35, Active: false},
		{Name: "Alice Brown", Email: "alice@example.com", Age: 28, Active: true},
	}

	ids, err := repo.InsertMany(ctx, users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created %d more users\n", len(ids))

	// ========== READ ==========
	fmt.Println("\n=== READ ===")

	// Find one user by ID
	foundUser, err := repo.FindOne(ctx, spec.Eq("_id", user.ID))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found user by ID: %s (%s)\n", foundUser.Name, foundUser.Email)

	// Find all active users
	activeUsers, err := repo.Find(ctx, spec.Eq("active", true))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d active users\n", len(activeUsers))

	// Find with sorting and limit
	fmt.Println("\nTop 2 oldest users:")
	oldestUsers, err := repo.Find(ctx, nil,
		repository.WithSort(bson.D{{"age", -1}}),
		repository.WithLimit(2),
	)
	if err != nil {
		log.Fatal(err)
	}
	for _, u := range oldestUsers {
		fmt.Printf("  - %s (age %d)\n", u.Name, u.Age)
	}

	// ========== UPDATE ==========
	fmt.Println("\n=== UPDATE ===")

	// Update one field
	matched, modified, err := repo.UpdateOne(ctx,
		spec.Eq("_id", user.ID),
		spec.Set("age", 31),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated user: matched=%d, modified=%d\n", matched, modified)

	// Update multiple fields
	_, _, err = repo.UpdateOne(ctx,
		spec.Eq("email", "bob@example.com"),
		spec.Combine(
			spec.Set("active", true),
			spec.Inc("age", 1),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated Bob: set active=true and incremented age")

	// Update many
	matched, modified, err = repo.UpdateMany(ctx,
		spec.Gte("age", 30),
		spec.Set("status", "senior"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated senior users: matched=%d, modified=%d\n", matched, modified)

	// ========== COUNT ==========
	fmt.Println("\n=== COUNT ===")

	total, err := repo.Count(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total users: %d\n", total)

	activeCount, err := repo.Count(ctx, spec.Eq("active", true))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Active users: %d\n", activeCount)

	// ========== DELETE ==========
	fmt.Println("\n=== DELETE ===")

	// Delete one
	deleted, err := repo.DeleteOne(ctx, spec.Eq("email", "bob@example.com"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted Bob: %d document(s)\n", deleted)

	// Delete many
	deleted, err = repo.DeleteMany(ctx, spec.Eq("active", false))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted inactive users: %d document(s)\n", deleted)

	// Final count
	remaining, _ := repo.Count(ctx, nil)
	fmt.Printf("\nRemaining users: %d\n", remaining)
}
