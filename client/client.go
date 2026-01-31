// Package client provides a MongoDB client wrapper with connection management
// and convenient methods for creating repositories.
//
// Example usage:
//
//	ctx := context.Background()
//	client, err := client.Connect(ctx, "mongodb://localhost:27017",
//	    client.WithDatabase("myapp"),
//	    client.WithMaxPoolSize(100),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close(ctx)
//
//	userRepo := client.Repository("users", User{})
package client

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Client wraps a MongoDB client with connection management and
// convenient repository creation methods.
type Client struct {
	client *mongo.Client
	db     *mongo.Database
	dbName string
}

// Option configures a Client.
type Option func(*clientOptions)

type clientOptions struct {
	database        string
	maxPoolSize     uint64
	minPoolSize     uint64
	maxConnIdleTime time.Duration
	connectTimeout  time.Duration
	serverTimeout   time.Duration
	appName         string
	directConn      bool
	replicaSet      string
}

// WithDatabase sets the default database name.
func WithDatabase(name string) Option {
	return func(o *clientOptions) {
		o.database = name
	}
}

// WithMaxPoolSize sets the maximum number of connections in the pool.
// Default is 100.
func WithMaxPoolSize(size uint64) Option {
	return func(o *clientOptions) {
		o.maxPoolSize = size
	}
}

// WithMinPoolSize sets the minimum number of connections in the pool.
// Default is 0.
func WithMinPoolSize(size uint64) Option {
	return func(o *clientOptions) {
		o.minPoolSize = size
	}
}

// WithMaxConnIdleTime sets the maximum time a connection can remain idle.
func WithMaxConnIdleTime(d time.Duration) Option {
	return func(o *clientOptions) {
		o.maxConnIdleTime = d
	}
}

// WithConnectTimeout sets the timeout for initial connection.
func WithConnectTimeout(d time.Duration) Option {
	return func(o *clientOptions) {
		o.connectTimeout = d
	}
}

// WithServerSelectionTimeout sets the timeout for server selection.
func WithServerSelectionTimeout(d time.Duration) Option {
	return func(o *clientOptions) {
		o.serverTimeout = d
	}
}

// WithAppName sets the application name for server logs.
func WithAppName(name string) Option {
	return func(o *clientOptions) {
		o.appName = name
	}
}

// WithDirectConnection enables direct connection mode.
// Use this when connecting to a single server without replica set.
func WithDirectConnection(direct bool) Option {
	return func(o *clientOptions) {
		o.directConn = direct
	}
}

// WithReplicaSet specifies the replica set name.
func WithReplicaSet(name string) Option {
	return func(o *clientOptions) {
		o.replicaSet = name
	}
}

// Connect creates a new MongoDB client and establishes a connection.
// The uri should be a valid MongoDB connection string.
//
// Example:
//
//	client, err := Connect(ctx, "mongodb://localhost:27017",
//	    WithDatabase("myapp"),
//	    WithMaxPoolSize(50),
//	)
func Connect(ctx context.Context, uri string, opts ...Option) (*Client, error) {
	// Apply options
	cfg := &clientOptions{
		database:       "test",
		maxPoolSize:    100,
		connectTimeout: 10 * time.Second,
		serverTimeout:  30 * time.Second,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Build mongo options
	mongoOpts := options.Client().ApplyURI(uri)

	if cfg.maxPoolSize > 0 {
		mongoOpts.SetMaxPoolSize(cfg.maxPoolSize)
	}
	if cfg.minPoolSize > 0 {
		mongoOpts.SetMinPoolSize(cfg.minPoolSize)
	}
	if cfg.maxConnIdleTime > 0 {
		mongoOpts.SetMaxConnIdleTime(cfg.maxConnIdleTime)
	}
	if cfg.connectTimeout > 0 {
		mongoOpts.SetConnectTimeout(cfg.connectTimeout)
	}
	if cfg.serverTimeout > 0 {
		mongoOpts.SetServerSelectionTimeout(cfg.serverTimeout)
	}
	if cfg.appName != "" {
		mongoOpts.SetAppName(cfg.appName)
	}
	if cfg.directConn {
		mongoOpts.SetDirect(true)
	}
	if cfg.replicaSet != "" {
		mongoOpts.SetReplicaSet(cfg.replicaSet)
	}

	// Connect
	client, err := mongo.Connect(ctx, mongoOpts)
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		client.Disconnect(ctx)
		return nil, err
	}

	return &Client{
		client: client,
		db:     client.Database(cfg.database),
		dbName: cfg.database,
	}, nil
}

// Close disconnects the MongoDB client and releases resources.
// Always call Close when done with the client.
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Ping verifies the connection to MongoDB.
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, readpref.Primary())
}

// Database returns a handle to the specified database.
// If no name is provided, returns the default database.
func (c *Client) Database(name ...string) *mongo.Database {
	if len(name) > 0 && name[0] != "" {
		return c.client.Database(name[0])
	}
	return c.db
}

// Collection returns a handle to the specified collection in the default database.
func (c *Client) Collection(name string) *mongo.Collection {
	return c.db.Collection(name)
}

// MongoClient returns the underlying mongo.Client for advanced operations.
func (c *Client) MongoClient() *mongo.Client {
	return c.client
}

// DatabaseName returns the default database name.
func (c *Client) DatabaseName() string {
	return c.dbName
}

// StartSession starts a new session for transaction support.
func (c *Client) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	return c.client.StartSession(opts...)
}

// UseSession creates a session and runs the provided function within it.
func (c *Client) UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return c.client.UseSession(ctx, fn)
}

// ListDatabaseNames returns a list of database names.
func (c *Client) ListDatabaseNames(ctx context.Context) ([]string, error) {
	return c.client.ListDatabaseNames(ctx, map[string]any{})
}

// ListCollectionNames returns a list of collection names in the default database.
func (c *Client) ListCollectionNames(ctx context.Context) ([]string, error) {
	return c.db.ListCollectionNames(ctx, map[string]any{})
}

// DropDatabase drops the specified database.
func (c *Client) DropDatabase(ctx context.Context, name string) error {
	return c.client.Database(name).Drop(ctx)
}

// DropCollection drops the specified collection from the default database.
func (c *Client) DropCollection(ctx context.Context, name string) error {
	return c.db.Collection(name).Drop(ctx)
}
