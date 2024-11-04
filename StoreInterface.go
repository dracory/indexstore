package indexstore

type StoreInterface interface {
	// SearchValueCreate(value *SearchValue) error
	// SearchValueDelete(value *SearchValue) error
	// SearchValueDeleteByID(valueID string) error
	// SearchValueFindByID(id string) (*SearchValue, error)
	// SearchValueFindBySourceReferenceID(sourceReferenceID string) (*SearchValue, error)
	// SearchValueList(options SearchValueQueryOptions) ([]SearchValue, error)
	// SearchValueSoftDelete(discount *SearchValue) error
	// SearchValueSoftDeleteByID(discountID string) error
	// SearchValueUpdate(value *SearchValue) error
	Truncate() error
}

// type SearchValueQueryOptions struct {
// 	ID                string
// 	IDIn              string
// 	SourceReferenceID string
// 	SearchValue       string
// 	SearchType        string
// 	Offset            int
// 	Limit             int
// 	SortOrder         string
// 	OrderBy           string
// 	CountOnly         bool
// 	WithDeleted       bool
// }
