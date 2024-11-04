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
// It will drop the table and create it again
func (st *Store) AutoMigrate() error {
	sqlStr := sb.NewBuilder(st.dbDriverName).Table(st.tableName).DropIfExists()

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := st.db.Exec(sqlStr)

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
	sqlStr := sb.NewBuilder(store.dbDriverName).Table(store.tableName).DropIfExists()

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := store.db.Exec(sqlStr)

	return err
}
