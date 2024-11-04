package indexstore

import (
	"database/sql"
	"log"

	"github.com/doug-martin/goqu/v9"
	"github.com/gouniverse/sb"
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

func (store *Store) Search(query SearchQuery) ([]map[string]any, error) {
	sqlStr, sqlParams, errSql := goqu.Dialect(store.dbDriverName).
		Select("*").
		Prepared(true).
		From(store.tableName).
		ToSQL()

	if errSql != nil {
		return nil, errSql
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	result, err := sb.NewDatabase(store.db, store.dbDriverName).SelectToMapAny(sqlStr, sqlParams...)

	return result, err
}
