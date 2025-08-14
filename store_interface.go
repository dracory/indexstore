package indexstore

type StoreInterface interface {
    // AutoMigrate ensures the index table exists and matches the configured schema
    AutoMigrate() error
    Insert(data map[string]any) error
    Search(query SearchQuery) ([]map[string]any, error)
    Drop() error
    Truncate() error
}
