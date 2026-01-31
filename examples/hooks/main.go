// Example: Using Lifecycle Hooks
//
// This example demonstrates how to use BeforeSave and AfterLoad hooks
// for validation, transformation, and computed fields.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dElCIoGio/mongox/document"
	mongorepo "github.com/dElCIoGio/mongox/repository/mongo"
	"github.com/dElCIoGio/mongox/spec"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User demonstrates BeforeSave for validation and normalization.
type User struct {
	document.Base `bson:",inline"`

	FirstName string `bson:"first_name"`
	LastName  string `bson:"last_name"`
	Email     string `bson:"email"`
	Age       int    `bson:"age"`

	// Computed field (set in AfterLoad)
	FullName string `bson:"-"` // Not stored in MongoDB
}

// BeforeSave validates and normalizes user data before saving.
// This is called automatically before InsertOne and ReplaceOne.
func (u *User) BeforeSave(ctx context.Context) error {
	// Normalize email to lowercase
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))

	// Trim whitespace from names
	u.FirstName = strings.TrimSpace(u.FirstName)
	u.LastName = strings.TrimSpace(u.LastName)

	// Validation
	if u.FirstName == "" {
		return errors.New("first name is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	if !strings.Contains(u.Email, "@") {
		return errors.New("email must be a valid email address")
	}
	if u.Age < 0 || u.Age > 150 {
		return errors.New("age must be between 0 and 150")
	}

	fmt.Printf("  [BeforeSave] Validated user: %s %s <%s>\n", u.FirstName, u.LastName, u.Email)
	return nil
}

// AfterLoad computes derived fields after loading from MongoDB.
// This is called automatically after FindOne and Find.
func (u *User) AfterLoad(ctx context.Context) error {
	// Compute full name
	u.FullName = strings.TrimSpace(u.FirstName + " " + u.LastName)

	fmt.Printf("  [AfterLoad] Computed FullName: %s\n", u.FullName)
	return nil
}

// Order demonstrates using hooks for business logic.
type Order struct {
	document.Base `bson:",inline"`

	CustomerID string    `bson:"customer_id"`
	Items      []string  `bson:"items"`
	Subtotal   float64   `bson:"subtotal"`
	Tax        float64   `bson:"tax"`
	Total      float64   `bson:"total"`
	Status     string    `bson:"status"`
	PlacedAt   time.Time `bson:"placed_at"`

	// Computed fields
	ItemCount int  `bson:"-"`
	IsPaid    bool `bson:"-"`
}

// BeforeSave calculates totals and sets defaults.
func (o *Order) BeforeSave(ctx context.Context) error {
	// Calculate tax (10%)
	o.Tax = o.Subtotal * 0.10

	// Calculate total
	o.Total = o.Subtotal + o.Tax

	// Set default status
	if o.Status == "" {
		o.Status = "pending"
	}

	// Set placed time
	if o.PlacedAt.IsZero() {
		o.PlacedAt = time.Now().UTC()
	}

	fmt.Printf("  [BeforeSave] Calculated order: subtotal=%.2f, tax=%.2f, total=%.2f\n",
		o.Subtotal, o.Tax, o.Total)

	return nil
}

// AfterLoad sets computed fields.
func (o *Order) AfterLoad(ctx context.Context) error {
	o.ItemCount = len(o.Items)
	o.IsPaid = o.Status == "paid"

	fmt.Printf("  [AfterLoad] Order loaded: %d items, paid=%v\n", o.ItemCount, o.IsPaid)
	return nil
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

	// Setup collections
	db := client.Database("example_db")
	db.Collection("users").Drop(ctx)
	db.Collection("orders").Drop(ctx)

	userRepo := mongorepo.New[User](db.Collection("users"))
	orderRepo := mongorepo.New[Order](db.Collection("orders"))

	// ========== USER HOOKS EXAMPLE ==========
	fmt.Println("=== USER HOOKS ===\n")

	// Create user with normalization
	fmt.Println("Creating user with unnormalized data:")
	user := &User{
		FirstName: "  John  ",
		LastName:  "Doe",
		Email:     "  JOHN@EXAMPLE.COM  ",
		Age:       30,
	}

	if err := userRepo.InsertOne(ctx, user); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User saved with normalized email: %s\n\n", user.Email)

	// Validation failure example
	fmt.Println("Attempting to create user with invalid data:")
	invalidUser := &User{
		FirstName: "",
		Email:     "invalid",
		Age:       200,
	}

	if err := userRepo.InsertOne(ctx, invalidUser); err != nil {
		fmt.Printf("Validation failed (expected): %s\n\n", err)
	}

	// Load user and compute FullName
	fmt.Println("Loading user from database:")
	loadedUser, err := userRepo.FindOne(ctx, spec.Eq("_id", user.ID))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Loaded user with computed FullName: %s\n\n", loadedUser.FullName)

	// ========== ORDER HOOKS EXAMPLE ==========
	fmt.Println("=== ORDER HOOKS ===\n")

	// Create order with automatic calculations
	fmt.Println("Creating order (tax and total calculated automatically):")
	order := &Order{
		CustomerID: user.ID.Hex(),
		Items:      []string{"Widget", "Gadget", "Gizmo"},
		Subtotal:   100.00,
	}

	if err := orderRepo.InsertOne(ctx, order); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Order created: subtotal=%.2f, tax=%.2f, total=%.2f, status=%s\n\n",
		order.Subtotal, order.Tax, order.Total, order.Status)

	// Load order and verify computed fields
	fmt.Println("Loading order from database:")
	loadedOrder, err := orderRepo.FindOne(ctx, spec.Eq("_id", order.ID))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Order loaded: ItemCount=%d, IsPaid=%v\n\n", loadedOrder.ItemCount, loadedOrder.IsPaid)

	// Update order status and reload
	fmt.Println("Updating order status to 'paid':")
	orderRepo.UpdateOne(ctx, spec.Eq("_id", order.ID), spec.Set("status", "paid"))

	paidOrder, _ := orderRepo.FindOne(ctx, spec.Eq("_id", order.ID))
	fmt.Printf("Reloaded order: IsPaid=%v\n", paidOrder.IsPaid)
}
