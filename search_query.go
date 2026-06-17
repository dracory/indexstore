package indexstore

// SearchQuery defines the query parameters for searching the index store
type SearchQuery struct {
	Where     map[string]any
	Offset    int
	Limit     int
	SortOrder string
	OrderBy   string
}
