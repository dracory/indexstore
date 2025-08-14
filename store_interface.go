package indexstore

// StoreInterface is the interface for the index store
type StoreInterface interface {
	// AutoMigrate ensures the index table exists and matches the configured schema
	AutoMigrate() error

	// Debug enables or disables SQL logging
	Debug(debug bool)

	// Insert inserts a record into the index table
	Insert(data map[string]any) error

	// InsertMany inserts multiple records into the index table
	InsertMany(data []map[string]any) error

	// Search queries the index table
	Search(query SearchQuery) ([]map[string]any, error)

	// Drop removes the index table
	Drop() error

	// Truncate removes all records from the index table
	Truncate() error
}
