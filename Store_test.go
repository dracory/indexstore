package indexstore

import (
	"database/sql"
	"os"
	"testing"

	"github.com/gouniverse/sb"
	_ "modernc.org/sqlite"
)

func initDB(filepath string) *sql.DB {
	os.Remove(filepath) // remove database
	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		panic(err)
	}

	return db
}

func Test_Store_WithAutoMigrate(t *testing.T) {
	db := initDB("test_store_with_automigrate.db")

	storeAutomigrateFalse, errAutomigrateFalse := NewStore(NewStoreOptions{
		TableName:          "test_index_with_automigrate_false",
		DB:                 db,
		AutomigrateEnabled: false,
		Columns: []sb.Column{
			{
				Name:          "id",
				Type:          sb.COLUMN_TYPE_INTEGER,
				PrimaryKey:    true,
				AutoIncrement: true,
			},
		},
	})

	if errAutomigrateFalse != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", errAutomigrateFalse.Error())
	}

	if storeAutomigrateFalse.automigrateEnabled != false {
		t.Fatalf("automigrateEnabled: Expected [false] received [%v]", storeAutomigrateFalse.automigrateEnabled)
	}

	storeAutomigrateTrue, errAutomigrateTrue := NewStore(NewStoreOptions{
		TableName:          "test_index_with_automigrate_true",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []sb.Column{
			{
				Name:          "id",
				Type:          sb.COLUMN_TYPE_INTEGER,
				PrimaryKey:    true,
				AutoIncrement: true,
			},
		},
	})

	if errAutomigrateTrue != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", errAutomigrateTrue.Error())
	}

	if storeAutomigrateTrue.automigrateEnabled != true {
		t.Fatalf("automigrateEnabled: Expected [true] received [%v]", storeAutomigrateTrue.automigrateEnabled)
	}
}

func Test_Store_Search(t *testing.T) {
	db := initDB("test_store_search.db")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_search",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []sb.Column{
			{
				Name:       "id",
				Type:       sb.COLUMN_TYPE_INTEGER,
				PrimaryKey: true,
			},
		},
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.Insert(map[string]any{
		"id": 1,
	})

	if err != nil {
		t.Fatalf("Insert: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.Insert(map[string]any{
		"id": 2,
	})

	if err != nil {
		t.Fatalf("Insert: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.Insert(map[string]any{
		"id": 3,
	})

	if err != nil {
		t.Fatalf("Insert: Expected [err] to be nill received [%v]", err.Error())
	}

	searchQuery := SearchQuery{
		Offset:      0,
		Limit:       10,
		SortOrder:   "asc",
		OrderBy:     "id",
		CountOnly:   false,
		WithDeleted: false,
	}

	result, err := store.Search(searchQuery)

	if err != nil {
		t.Fatalf("Search: Expected [err] to be nill received [%v]", err.Error())
	}

	if len(result) != 3 {
		t.Fatalf("Search: Expected [3] received [%v]", len(result))
	}
}
