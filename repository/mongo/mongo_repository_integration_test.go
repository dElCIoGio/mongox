//go:build integration

package mongorepo_test

import (
	"context"
	"testing"
	"time"

	"github.com/dElCIoGio/mongox/document"
	mongorepo "github.com/dElCIoGio/mongox/repository/mongo"
	mongospec "github.com/dElCIoGio/mongox/spec"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mopt "go.mongodb.org/mongo-driver/mongo/options"

	mongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
)

type Order struct {
	document.Base `bson:",inline"`

	TenantID string `bson:"tenant_id"`
	Paid     bool   `bson:"paid"`
	Total    int    `bson:"total"`

	// hook flags
	BeforeSaveCalled bool `bson:"-"`
	AfterLoadCalled  bool `bson:"-"`
}

func (o *Order) BeforeSave(ctx context.Context) error {
	o.BeforeSaveCalled = true
	// Example validation: total must be non-negative
	if o.Total < 0 {
		return mongo.ErrClientDisconnected // any error; just to prove it stops insert
	}
	return nil
}

func (o *Order) AfterLoad(ctx context.Context) error {
	o.AfterLoadCalled = true
	return nil
}

func setupMongo(t *testing.T) (*mongo.Client, func()) {
	t.Helper()

	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("start mongodb container: %v", err)
	}

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("get mongodb uri: %v", err)
	}

	client, err := mongo.Connect(ctx, mopt.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}

	cleanup := func() {
		_ = client.Disconnect(ctx)
		_ = container.Terminate(ctx)
	}

	return client, cleanup
}

func TestInsertOne_AutoTouchAndBeforeSaveHook(t *testing.T) {
	client, cleanup := setupMongo(t)
	defer cleanup()

	ctx := context.Background()
	coll := client.Database("testdb").Collection("orders_insert")

	repo := mongorepo.New[Order](coll)

	doc := &Order{
		TenantID: "t1",
		Paid:     true,
		Total:    120,
	}

	if err := repo.InsertOne(ctx, doc); err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Auto-touch checks
	if doc.ID.IsZero() {
		t.Fatal("expected Base.ID to be set")
	}
	if doc.CreatedAt.IsZero() {
		t.Fatal("expected Base.CreatedAt to be set")
	}
	if doc.UpdatedAt.IsZero() {
		t.Fatal("expected Base.UpdatedAt to be set")
	}

	// Hook check
	if !doc.BeforeSaveCalled {
		t.Fatal("expected BeforeSave to be called")
	}

	// Confirm persisted
	count, err := coll.CountDocuments(ctx, bson.M{"_id": doc.ID})
	if err != nil {
		t.Fatalf("CountDocuments: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 doc, got %d", count)
	}
}

func TestFindOne_AfterLoadHook(t *testing.T) {
	client, cleanup := setupMongo(t)
	defer cleanup()

	ctx := context.Background()
	coll := client.Database("testdb").Collection("orders_findone")

	repo := mongorepo.New[Order](coll)

	doc := &Order{
		TenantID: "t1",
		Paid:     true,
		Total:    200,
	}

	if err := repo.InsertOne(ctx, doc); err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	got, err := repo.FindOne(ctx, mongospec.Eq("_id", doc.ID))
	if err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	if got == nil {
		t.Fatal("expected a document")
	}
	if got.ID != doc.ID {
		t.Fatal("expected same ID")
	}
	if !got.AfterLoadCalled {
		t.Fatal("expected AfterLoad to be called")
	}
}

func TestUpdateOne_InjectsUpdatedAtIntoSet(t *testing.T) {
	client, cleanup := setupMongo(t)
	defer cleanup()

	ctx := context.Background()
	coll := client.Database("testdb").Collection("orders_update")

	repo := mongorepo.New[Order](coll)

	doc := &Order{
		TenantID: "t1",
		Paid:     false,
		Total:    50,
	}

	if err := repo.InsertOne(ctx, doc); err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	createdAt := doc.CreatedAt
	oldUpdatedAt := doc.UpdatedAt

	// Small sleep to ensure updated_at moves forward
	time.Sleep(10 * time.Millisecond)

	matched, modified, err := repo.UpdateOne(ctx,
		mongospec.Eq("_id", doc.ID),
		bson.M{"$set": bson.M{"paid": true}},
	)
	if err != nil {
		t.Fatalf("UpdateOne failed: %v", err)
	}
	if matched != 1 || modified != 1 {
		t.Fatalf("expected matched=1 modified=1, got matched=%d modified=%d", matched, modified)
	}

	got, err := repo.FindOne(ctx, mongospec.Eq("_id", doc.ID))
	if err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	if !got.Paid {
		t.Fatal("expected paid=true after update")
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Fatal("expected created_at unchanged")
	}
	if !got.UpdatedAt.After(oldUpdatedAt) {
		t.Fatalf("expected updated_at to increase, old=%v new=%v", oldUpdatedAt, got.UpdatedAt)
	}
}

func TestReplaceOne_TouchesUpdatedAtAndCallsBeforeSave(t *testing.T) {
	client, cleanup := setupMongo(t)
	defer cleanup()

	ctx := context.Background()
	coll := client.Database("testdb").Collection("orders_replace")

	repo := mongorepo.New[Order](coll)

	doc := &Order{
		TenantID: "t1",
		Paid:     false,
		Total:    10,
	}
	if err := repo.InsertOne(ctx, doc); err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	oldUpdatedAt := doc.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	replacement := &Order{
		// keep same ID, replace rest
		Base: document.Base{
			ID:        doc.ID,
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		},
		TenantID: "t1",
		Paid:     true,
		Total:    999,
	}

	matched, modified, err := repo.ReplaceOne(ctx, mongospec.Eq("_id", doc.ID), replacement)
	if err != nil {
		t.Fatalf("ReplaceOne failed: %v", err)
	}
	if matched != 1 {
		t.Fatalf("expected matched=1 got %d", matched)
	}

	if !replacement.BeforeSaveCalled {
		t.Fatal("expected BeforeSave called on replacement doc")
	}

	got, err := repo.FindOne(ctx, mongospec.Eq("_id", doc.ID))
	if err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	if got.Total != 999 || !got.Paid {
		t.Fatal("expected replaced fields to persist")
	}
	if !got.UpdatedAt.After(oldUpdatedAt) {
		t.Fatalf("expected updated_at to increase, old=%v new=%v", oldUpdatedAt, got.UpdatedAt)
	}
}
