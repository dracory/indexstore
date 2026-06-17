package indexstore

import (
	"database/sql"
)

// ColumnDefinition defines a column for the index table
type ColumnDefinition struct {
	Name       string
	Type       string
	PrimaryKey bool
}

// NewStoreOptions define the options for creating a new index store
type NewStoreOptions struct {
	TableName          string
	DB                 *sql.DB
	AutomigrateEnabled bool
	DebugEnabled       bool
	Columns            []ColumnDefinition
}
