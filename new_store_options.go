package indexstore

import (
	"database/sql"

	"github.com/gouniverse/sb"
)

// NewStoreOptions define the options for creating a new session store
type NewStoreOptions struct {
	TableName          string
	DB                 *sql.DB
	DbDriverName       string
	AutomigrateEnabled bool
	DebugEnabled       bool
	Columns            []sb.Column
}
