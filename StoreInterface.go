package indexstore

type StoreInterface interface {
	Insert(data map[string]any) error
	Search(query SearchQuery) ([]map[string]any, error)
	Drop() error
	Truncate() error
}
