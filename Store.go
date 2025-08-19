package indexstore

import (
	"database/sql"
	"errors"
	"log"

	// "strconv"

	"github.com/doug-martin/goqu/v9"
	"github.com/gouniverse/sb"
	"github.com/spf13/cast"
)

var _ StoreInterface = (*Store)(nil) // verify it extends the interface

// Store is the index store
type Store struct {
	tableName          string
	db                 *sql.DB
	dbDriverName       string
	automigrateEnabled bool
	debugEnabled       bool
	columns            []sb.Column
}

// AutoMigrate auto migrate
//
// It will drop the index table and create it again
func (st *Store) AutoMigrate() error {
	sql := st.sqlTableCreate()

	if st.debugEnabled {
		log.Println(sql)
	}

	_, err := st.db.Exec(sql)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// EnableDebug - enables the debug option
func (st *Store) EnableDebug(debug bool) {
	st.debugEnabled = debug
}

// Debug enables or disables SQL logging (alias of EnableDebug)
func (st *Store) Debug(debug bool) {
	st.debugEnabled = debug
}

func (store *Store) Insert(data map[string]any) error {
	sqlStr, _, errSql := goqu.Dialect(store.dbDriverName).
		Insert(store.tableName).
		Rows(data).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := store.db.Exec(sqlStr)

	return err
}

func (store *Store) InsertMany(data []map[string]any) error {
	sqlStr, _, errSql := goqu.Dialect(store.dbDriverName).
		Insert(store.tableName).
		Rows(data).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := store.db.Exec(sqlStr)

	return err
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
	where := goqu.Ex{}
	for k, v := range filters {
		where[k] = v
	}

	ds := goqu.Dialect(store.dbDriverName).
		Delete(store.tableName)

	if len(where) > 0 {
		ds = ds.Where(where)
	}

	sqlStr, params, errSql := ds.ToSQL()
	if errSql != nil {
		return errSql
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := sb.NewDatabase(store.db, store.dbDriverName).Exec(sqlStr, params...)
	return err
}

// UpsertWhereEquals performs an upsert using equality filters as the match criteria.
// It updates non-filter columns when a match exists, otherwise inserts a new row
// composed from filters and data (data keys override filters on conflict).
func (store *Store) UpsertWhereEquals(filters map[string]any, data map[string]any) error {
	// Build update map with keys from data excluding filter keys
	updateMap := goqu.Record{}
	for k, v := range data {
		if _, isFilter := filters[k]; isFilter {
			continue
		}
		updateMap[k] = v
	}

	// Attempt UPDATE if there is anything to update
	if len(updateMap) > 0 {
		sqlStr, params, errSql := goqu.Dialect(store.dbDriverName).
			Update(store.tableName).
			Set(updateMap).
			Where(goqu.Ex(filters)).
			ToSQL()

		if errSql != nil {
			return errSql
		}

		if store.debugEnabled {
			log.Println(sqlStr)
		}

		res, err := sb.NewDatabase(store.db, store.dbDriverName).Exec(sqlStr, params...)
		if err != nil {
			return err
		}
		if res != nil {
			if affected, _ := res.RowsAffected(); affected > 0 {
				return nil
			}
		}
	} else {
		// No fields to update; if a row exists, we are done.
		if count, err := store.Count(SearchQuery{Where: goqu.Ex(filters)}); err == nil && count > 0 {
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
	sqlStr, _, errSql := goqu.Dialect(store.dbDriverName).
		Truncate(store.tableName).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := store.db.Exec(sqlStr)

	return err
}

func (store *Store) Drop() error {
	sqlStr := sb.NewBuilder(store.dbDriverName).
		Table(store.tableName).
		DropIfExists()

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := store.db.Exec(sqlStr)

	return err
}

func (store *Store) query(query SearchQuery) *goqu.SelectDataset {
	q := goqu.Dialect(store.dbDriverName).
		Select("*").
		Prepared(true).
		From(store.tableName)

	if query.Where != nil {
		q = q.Where(query.Where)
	}

	if query.OrderBy != "" {
		if query.SortOrder == "asc" {
			q = q.Order(goqu.I(query.OrderBy).Asc())
		} else {
			q = q.Order(goqu.I(query.OrderBy).Desc())
		}
	}

	if query.Offset > 0 {
		q = q.Offset(cast.ToUint(query.Offset))
	}

	if query.Limit > 0 {
		q = q.Limit(cast.ToUint(query.Limit))
	}

	return q
}

func (store *Store) Search(query SearchQuery) ([]map[string]any, error) {
	q := store.query(query)

	sqlStr, sqlParams, errSql := q.ToSQL()

	if errSql != nil {
		return nil, errSql
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	result, err := sb.NewDatabase(store.db, store.dbDriverName).SelectToMapAny(sqlStr, sqlParams...)

	return result, err
}

func (store *Store) Count(query SearchQuery) (int64, error) {
	q := store.query(query)

	sqlStr, params, errSql := q.Prepared(true).
		Limit(1).
		Select(goqu.COUNT(goqu.Star()).As("count")).
		ToSQL()

	if errSql != nil {
		return -1, nil
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	db := sb.NewDatabase(store.db, store.dbDriverName)
	mapped, err := db.SelectToMapString(sqlStr, params...)
	if err != nil {
		return -1, err
	}

	if len(mapped) < 1 {
		return -1, nil
	}

	countStr := mapped[0]["count"]

	i, err := cast.ToInt64E(countStr)

	if err != nil {
		return -1, err

	}

	return i, nil
}
