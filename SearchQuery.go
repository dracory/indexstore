package indexstore

import "github.com/doug-martin/goqu/v9"

type SearchQuery struct {
	Where     goqu.Expression
	Offset    int
	Limit     int
	SortOrder string
	OrderBy   string
}
