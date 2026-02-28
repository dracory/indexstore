package indexstore

import (
	"errors"

	"github.com/dracory/sb"
)

// NewStore creates a new entity store
func NewStore(opts NewStoreOptions) (*Store, error) {
	store := &Store{
		tableName:          opts.TableName,
		automigrateEnabled: opts.AutomigrateEnabled,
		db:                 opts.DB,
		dbDriverName:       opts.DbDriverName,
		debugEnabled:       opts.DebugEnabled,
		columns:            opts.Columns,
	}

	if store.tableName == "" {
		return nil, errors.New("index store: TableName is required")
	}

	if store.db == nil {
		return nil, errors.New("index store: DB is required")
	}

	if store.columns == nil {
		return nil, errors.New("index store: Columns is required")
	}

	if store.dbDriverName == "" {
		store.dbDriverName = sb.DatabaseDriverName(store.db)
	}

	if len(store.columns) < 1 {
		return nil, errors.New("index store: Columns number must be greater than 0")
	}

	if store.automigrateEnabled {
		err := store.AutoMigrate()

		if err != nil {
			return nil, err
		}
	}

	return store, nil
}
