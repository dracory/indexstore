package indexstore

import (
	"database/sql"
	"os"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/gouniverse/sb"
	"github.com/spf13/cast"
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
				Name:       "id",
				Type:       sb.COLUMN_TYPE_INTEGER,
				PrimaryKey: true,
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

func Test_Store_Insert(t *testing.T) {
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
}

func Test_Store_InsertMany(t *testing.T) {
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

	err = store.InsertMany([]map[string]any{
		{
			"id": 1,
		},
		{
			"id": 2,
		},
		{
			"id": 3,
		},
	})

	if err != nil {
		t.Fatalf("InsertMany: Expected [err] to be nill received [%v]", err.Error())
	}
}

func Test_Store_Search_SelectAll(t *testing.T) {
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

	err = store.InsertMany([]map[string]any{
		{
			"id": 1,
		},
		{
			"id": 2,
		},
		{
			"id": 3,
		},
	})

	if err != nil {
		t.Fatalf("Insert: Expected [err] to be nill received [%v]", err.Error())
	}

	searchQuery := SearchQuery{
		Offset:    0,
		Limit:     10,
		SortOrder: "asc",
		OrderBy:   "id",
	}

	result, err := store.Search(searchQuery)

	if err != nil {
		t.Fatalf("Search: Expected [err] to be nill received [%v]", err.Error())
	}

	if len(result) != 3 {
		t.Fatalf("Search: Expected [3] received [%v]", len(result))
	}

	if cast.ToInt(result[0]["id"]) != 1 {
		t.Fatalf("Search: Expected [1] received [%v]", result[0]["id"])
	}
}

func Test_Store_Search_Where(t *testing.T) {
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
			{
				Name: "name",
				Type: sb.COLUMN_TYPE_TEXT,
			},
		},
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.InsertMany([]map[string]any{
		{
			"id":   1,
			"name": "Alice",
		},
		{
			"id":   2,
			"name": "Bob",
		},
		{
			"id":   3,
			"name": "Charlie",
		},
	})

	if err != nil {
		t.Fatalf("Insert: Expected [err] to be nill received [%v]", err.Error())
	}

	searchQuery := SearchQuery{
		Where:     goqu.C("id").Eq(2),
		Offset:    0,
		Limit:     10,
		SortOrder: "asc",
		OrderBy:   "id",
	}

	result, err := store.Search(searchQuery)

	if err != nil {
		t.Fatalf("Search: Expected [err] to be nill received [%v]", err.Error())
	}

	if len(result) != 1 {
		t.Fatalf("Search: Expected [3] received [%v]", len(result))
	}

	if cast.ToInt(result[0]["id"]) != 2 {
		t.Fatalf("Search: Expected [2] received [%v]", result[0]["id"])
	}

	if cast.ToString(result[0]["name"]) != "Bob" {
		t.Fatalf("Search: Expected [Bob] received [%v]", result[0]["name"])
	}
}

func Test_Store_Search_WhereExpression(t *testing.T) {
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
			{
				Name: "name",
				Type: sb.COLUMN_TYPE_TEXT,
			},
		},
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.InsertMany([]map[string]any{
		{
			"id":   1,
			"name": "Alice",
		},
		{
			"id":   2,
			"name": "Bob",
		},
		{
			"id":   3,
			"name": "Charlie",
		},
	})

	if err != nil {
		t.Fatalf("Insert: Expected [err] to be nill received [%v]", err.Error())
	}

	searchQuery := SearchQuery{
		Where: goqu.Or(
			goqu.Ex{
				"id": goqu.Op{"eq": 2},
				// "a": goqu.Op{"gt": 10},
				// "b": goqu.Op{"lt": 10},
			},
			goqu.Ex{
				"id": goqu.Op{"eq": 3},
				// "c": nil,
				// "d": []string{"a", "b", "c"},
			},
		),
		Offset:    0,
		Limit:     10,
		SortOrder: "asc",
		OrderBy:   "id",
	}

	result, err := store.Search(searchQuery)

	if err != nil {
		t.Fatalf("Search: Expected [err] to be nill received [%v]", err.Error())
	}

	if len(result) != 2 {
		t.Fatalf("Search: Expected [3] received [%v]", len(result))
	}

	if cast.ToInt(result[0]["id"]) != 2 {
		t.Fatalf("Search: Expected [2] received [%v]", result[0]["id"])
	}

	if cast.ToString(result[0]["name"]) != "Bob" {
		t.Fatalf("Search: Expected [Bob] received [%v]", result[0]["name"])
	}
}

func Test_Store_Count(t *testing.T) {
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

	err = store.InsertMany([]map[string]any{
		{
			"id": 1,
		},
		{
			"id": 2,
		},
		{
			"id": 3,
		},
	})

	if err != nil {
		t.Fatalf("Insert: Expected [err] to be nill received [%v]", err.Error())
	}

	searchQuery := SearchQuery{
		Where: goqu.Or(
			goqu.Ex{
				"id": goqu.Op{"eq": 2},
			},
			goqu.Ex{
				"id": goqu.Op{"eq": 3},
			},
		),
		// Offset:    0,
		// Limit:     10,
		SortOrder: "asc",
		OrderBy:   "id",
	}
	count, err := store.Count(searchQuery)

	if err != nil {
		t.Fatalf("Count: Expected [err] to be nill received [%v]", err.Error())
	}

	if count != 2 {
		t.Fatalf("Count: Expected [3] received [%v]", count)
	}

}
