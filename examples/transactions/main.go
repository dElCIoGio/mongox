// Example: Using Transactions
//
// This example demonstrates how to use MongoDB transactions to ensure
// atomic operations across multiple documents and collections.
//
// Note: Transactions require a MongoDB replica set. For local development,
// you can use: mongod --replSet rs0
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dElCIoGio/mongox/document"
	mongorepo "github.com/dElCIoGio/mongox/repository/mongo"
	"github.com/dElCIoGio/mongox/spec"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Account represents a bank account.
type Account struct {
	document.Base `bson:",inline"`

	OwnerName string  `bson:"owner_name"`
	Balance   float64 `bson:"balance"`
}

// Transaction represents a money transfer record.
type Transaction struct {
	document.Base `bson:",inline"`

	FromAccountID string    `bson:"from_account_id"`
	ToAccountID   string    `bson:"to_account_id"`
	Amount        float64   `bson:"amount"`
	Status        string    `bson:"status"`
	Timestamp     time.Time `bson:"timestamp"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to MongoDB (must be a replica set for transactions)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Setup collections
	db := client.Database("example_db")
	db.Collection("accounts").Drop(ctx)
	db.Collection("transactions").Drop(ctx)

	accountRepo := mongorepo.New[Account](db.Collection("accounts"))
	txnRepo := mongorepo.New[Transaction](db.Collection("transactions"))

	// Create test accounts
	alice := &Account{OwnerName: "Alice", Balance: 1000.00}
	bob := &Account{OwnerName: "Bob", Balance: 500.00}

	accountRepo.InsertOne(ctx, alice)
	accountRepo.InsertOne(ctx, bob)

	fmt.Println("Initial balances:")
	fmt.Printf("  Alice: $%.2f\n", alice.Balance)
	fmt.Printf("  Bob: $%.2f\n\n", bob.Balance)

	// Create transaction manager
	tm := mongorepo.NewTransactionManager(client, nil)

	// ========== SUCCESSFUL TRANSFER ==========
	fmt.Println("=== Successful Transfer ===\n")

	err = transferMoney(ctx, tm, accountRepo, txnRepo, alice.ID.Hex(), bob.ID.Hex(), 200.00)
	if err != nil {
		fmt.Printf("Transfer failed: %s\n", err)
	} else {
		fmt.Println("Transfer of $200 from Alice to Bob succeeded!\n")
	}

	// Verify balances
	alice, _ = accountRepo.FindOne(ctx, spec.Eq("_id", alice.ID))
	bob, _ = accountRepo.FindOne(ctx, spec.Eq("_id", bob.ID))
	fmt.Println("After successful transfer:")
	fmt.Printf("  Alice: $%.2f\n", alice.Balance)
	fmt.Printf("  Bob: $%.2f\n\n", bob.Balance)

	// ========== FAILED TRANSFER (INSUFFICIENT FUNDS) ==========
	fmt.Println("=== Failed Transfer (Insufficient Funds) ===\n")

	// Try to transfer more than Alice has
	err = transferMoney(ctx, tm, accountRepo, txnRepo, alice.ID.Hex(), bob.ID.Hex(), 1000.00)
	if err != nil {
		fmt.Printf("Transfer failed (expected): %s\n\n", err)
	}

	// Verify balances unchanged
	alice, _ = accountRepo.FindOne(ctx, spec.Eq("_id", alice.ID))
	bob, _ = accountRepo.FindOne(ctx, spec.Eq("_id", bob.ID))
	fmt.Println("After failed transfer (balances unchanged):")
	fmt.Printf("  Alice: $%.2f\n", alice.Balance)
	fmt.Printf("  Bob: $%.2f\n\n", bob.Balance)

	// ========== USING CONVENIENCE FUNCTION ==========
	fmt.Println("=== Using Convenience Function ===\n")

	err = mongorepo.RunInTransaction(ctx, client, func(txCtx context.Context) error {
		// Multiple operations in one transaction
		_, _, err := accountRepo.UpdateOne(txCtx, spec.Eq("_id", alice.ID), spec.Inc("balance", 50))
		if err != nil {
			return err
		}
		_, _, err = accountRepo.UpdateOne(txCtx, spec.Eq("_id", bob.ID), spec.Inc("balance", 50))
		return err
	})

	if err != nil {
		fmt.Printf("Batch update failed: %s\n", err)
	} else {
		fmt.Println("Batch deposit of $50 to both accounts succeeded!\n")
	}

	// Final balances
	alice, _ = accountRepo.FindOne(ctx, spec.Eq("_id", alice.ID))
	bob, _ = accountRepo.FindOne(ctx, spec.Eq("_id", bob.ID))
	fmt.Println("Final balances:")
	fmt.Printf("  Alice: $%.2f\n", alice.Balance)
	fmt.Printf("  Bob: $%.2f\n\n", bob.Balance)

	// Show transaction log
	txns, _ := txnRepo.Find(ctx, nil)
	fmt.Printf("Transaction log: %d transactions recorded\n", len(txns))
	for _, t := range txns {
		fmt.Printf("  - $%.2f from %s to %s [%s]\n",
			t.Amount, t.FromAccountID[:8], t.ToAccountID[:8], t.Status)
	}
}

// transferMoney performs an atomic money transfer between two accounts.
// If any step fails, the entire operation is rolled back.
func transferMoney(
	ctx context.Context,
	tm *mongorepo.MongoTransactionManager,
	accountRepo *mongorepo.MongoRepository[Account],
	txnRepo *mongorepo.MongoRepository[Transaction],
	fromID, toID string,
	amount float64,
) error {
	return tm.WithTransaction(ctx, func(txCtx context.Context) error {
		// 1. Get source account
		from, err := accountRepo.FindOne(txCtx, spec.Eq("_id", fromID))
		if err != nil {
			return fmt.Errorf("source account not found: %w", err)
		}

		// 2. Check sufficient funds
		if from.Balance < amount {
			return errors.New("insufficient funds")
		}

		// 3. Debit source account
		_, _, err = accountRepo.UpdateOne(txCtx,
			spec.Eq("_id", fromID),
			spec.Inc("balance", -amount),
		)
		if err != nil {
			return fmt.Errorf("failed to debit account: %w", err)
		}

		// 4. Credit destination account
		_, _, err = accountRepo.UpdateOne(txCtx,
			spec.Eq("_id", toID),
			spec.Inc("balance", amount),
		)
		if err != nil {
			return fmt.Errorf("failed to credit account: %w", err)
		}

		// 5. Record transaction
		txn := &Transaction{
			FromAccountID: fromID,
			ToAccountID:   toID,
			Amount:        amount,
			Status:        "completed",
			Timestamp:     time.Now().UTC(),
		}
		if err := txnRepo.InsertOne(txCtx, txn); err != nil {
			return fmt.Errorf("failed to record transaction: %w", err)
		}

		return nil
	})
}
