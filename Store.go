package indexstore

import (
	"database/sql"
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
	err := st.Drop()

	if err != nil {
		return err
	}

	sql := st.sqlTableCreate()

	if st.debugEnabled {
		log.Println(sql)
	}

	_, err = st.db.Exec(sql)

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
