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

	// Upsert inserts or updates a record based on the specified conflict columns (e.g., unique keys)
	// When a conflict occurs, columns present in data (except the conflict columns) will be updated
	Upsert(data map[string]any, conflictColumns []string) error

	// DeleteWhereEquals deletes rows matching all equality filters provided
	// Example: DeleteWhereEquals(map[string]any{"record_id": id, "status": "draft"})
	DeleteWhereEquals(filters map[string]any) error

	// UpsertWhereEquals performs an upsert using equality filters as the match criteria.
	// It updates non-filter columns when a match exists, otherwise inserts a new row
	// composed from filters and data (data keys override filters on conflict).
	UpsertWhereEquals(filters map[string]any, data map[string]any) error

	// Search queries the index table
	Search(query SearchQuery) ([]map[string]any, error)

	// Drop removes the index table
	Drop() error

	// Truncate removes all records from the index table
	Truncate() error
}
