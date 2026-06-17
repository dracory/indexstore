package indexstore

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/dracory/neat"
)

// NewStore creates a new index store
func NewStore(opts NewStoreOptions) (*Store, error) {
	if opts.DB == nil {
		return nil, errors.New("index store: DB is required")
	}

	if opts.TableName == "" {
		return nil, errors.New("index store: TableName is required")
	}

	neatDB, err := neat.NewFromSQLDB(opts.DB)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	store := &Store{
		tableName:          opts.TableName,
		automigrateEnabled: opts.AutomigrateEnabled,
		db:                 neatDB,
		debugEnabled:       opts.DebugEnabled,
		logger:             logger,
		columns:            opts.Columns,
	}

	if store.automigrateEnabled {
		err := store.MigrateUp(context.Background())
		if err != nil {
			return nil, err
		}
	}

	return store, nil
}
