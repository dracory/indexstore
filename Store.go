package indexstore

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"

	"github.com/dracory/neat"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
)

var _ StoreInterface = (*Store)(nil) // verify it extends the interface

// Store is the index store
type Store struct {
	tableName          string
	db                 *neat.Database
	automigrateEnabled bool
	debugEnabled       bool
	logger             *slog.Logger
	columns            []ColumnDefinition
}

// MigrateUp creates the index table
func (st *Store) MigrateUp(ctx context.Context, tx ...*sql.Tx) error {
	if st.db.Schema().HasTable(st.tableName) {
		if st.debugEnabled {
			st.logger.Info("MigrateUp: table already exists", "table", st.tableName)
		}
		return nil
	}

	err := st.db.Schema().Create(st.tableName, func(table contractsschema.Blueprint) {
		if len(st.columns) == 0 {
			// Default columns if none provided
			table.String("id", 21)
			table.Primary("id")
			table.Text("data")
			table.DateTime("created_at")
			table.DateTime("updated_at")
		} else {
			// Use provided column definitions
			for _, col := range st.columns {
				switch col.Type {
				case "string":
					table.String(col.Name, 255)
				case "text":
					table.Text(col.Name)
				case "integer":
					table.Integer(col.Name)
				case "datetime":
					table.DateTime(col.Name)
				default:
					table.String(col.Name, 255)
				}
				if col.PrimaryKey {
					table.Primary(col.Name)
				}
			}
		}
	})

	if err != nil {
		if st.debugEnabled {
			st.logger.Error("MigrateUp failed", "error", err)
		}
		return err
	}

	return nil
}

// MigrateDown drops the index table
func (st *Store) MigrateDown(ctx context.Context, tx ...*sql.Tx) error {
	if !st.db.Schema().HasTable(st.tableName) {
		if st.debugEnabled {
			st.logger.Info("MigrateDown: table does not exist", "table", st.tableName)
		}
		return nil
	}

	err := st.db.Schema().Drop(st.tableName)
	if err != nil {
		if st.debugEnabled {
			st.logger.Error("MigrateDown failed", "error", err)
		}
		return err
	}
	return nil
}

// EnableDebug - enables the debug option
func (st *Store) EnableDebug(debug bool) {
	st.debugEnabled = debug
	if debug {
		st.db.EnableDebug()
		st.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		st.db.DisableDebug()
		st.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
}

// Debug enables or disables SQL logging (alias of EnableDebug)
func (st *Store) Debug(debug bool) {
	st.EnableDebug(debug)
}

func (store *Store) Insert(data map[string]any) error {
	return store.db.Query().Table(store.tableName).Create(data)
}

func (store *Store) InsertMany(data []map[string]any) error {
	for _, row := range data {
		if err := store.db.Query().Table(store.tableName).Create(row); err != nil {
			return err
		}
	}
	return nil
}

// Upsert keeps backward compatibility with the original interface by
// delegating to UpsertWhereEquals using the provided conflict columns
func (store *Store) Upsert(data map[string]any, conflictColumns []string) error {
	filters := map[string]any{}
	for _, col := range conflictColumns {
		val, ok := data[col]
		if !ok {
			return errors.New("upsert: missing conflict column value for " + col)
		}
		filters[col] = val
	}
	return store.UpsertWhereEquals(filters, data)
}

// DeleteWhereEquals deletes rows matching all equality filters.
// Example: DeleteWhereEquals(map[string]any{"record_id": id, "status": "draft"})
func (store *Store) DeleteWhereEquals(filters map[string]any) error {
	whereClause := ""
	args := []any{}
	for k, v := range filters {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += k + " = ?"
		args = append(args, v)
	}
	_, err := store.db.Query().Table(store.tableName).Where(whereClause, args...).Delete()
	return err
}

// UpsertWhereEquals performs an upsert using equality filters as the match criteria.
// It updates non-filter columns when a match exists, otherwise inserts a new row
// composed from filters and data (data keys override filters on conflict).
func (store *Store) UpsertWhereEquals(filters map[string]any, data map[string]any) error {
	// Build update map with keys from data excluding filter keys
	updateMap := map[string]any{}
	for k, v := range data {
		if _, isFilter := filters[k]; isFilter {
			continue
		}
		updateMap[k] = v
	}

	// Attempt UPDATE if there is anything to update
	if len(updateMap) > 0 {
		whereClause := ""
		args := []any{}
		for k, v := range filters {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += k + " = ?"
			args = append(args, v)
		}
		_, err := store.db.Query().Table(store.tableName).Where(whereClause, args...).Update(updateMap)
		if err != nil {
			return err
		}
		// Check if row exists
		count, err := store.Count(SearchQuery{Where: filters})
		if err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
	} else {
		// No fields to update; if a row exists, we are done.
		if count, err := store.Count(SearchQuery{Where: filters}); err == nil && count > 0 {
			return nil
		}
	}

	// Prepare insert row by merging filters and data (data overrides)
	row := map[string]any{}
	for k, v := range filters {
		row[k] = v
	}
	for k, v := range data {
		row[k] = v
	}
	return store.Insert(row)
}

func (store *Store) Truncate() error {
	_, err := store.db.Query().Table(store.tableName).Delete()
	return err
}

func (store *Store) Drop() error {
	err := store.db.Schema().Drop(store.tableName)
	return err
}

func (store *Store) query(query SearchQuery) contractsorm.Query {
	q := store.db.Query().Table(store.tableName)

	if query.Where != nil {
		whereClause := ""
		args := []any{}
		for k, v := range query.Where {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += k + " = ?"
			args = append(args, v)
		}
		q = q.Where(whereClause, args...)
	}

	if query.OrderBy != "" {
		if query.SortOrder == "asc" {
			q = q.OrderBy(query.OrderBy, "asc")
		} else {
			q = q.OrderBy(query.OrderBy, "desc")
		}
	}

	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}

	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}

	return q
}

func (store *Store) Search(query SearchQuery) ([]map[string]any, error) {
	q := store.query(query)

	var results []map[string]any
	err := q.Get(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (store *Store) Count(query SearchQuery) (int64, error) {
	q := store.query(query)

	var count int64
	err := q.Count(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}
