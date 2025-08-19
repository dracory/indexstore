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

func Test_Store_DeleteWhereEquals(t *testing.T) {
    db := initDB("test_store_delete_where_equals.db")

    store, err := NewStore(NewStoreOptions{
        TableName:          "test_index_delete",
        DB:                 db,
        AutomigrateEnabled: true,
        Columns: []sb.Column{
            {Name: "id", Type: sb.COLUMN_TYPE_INTEGER, PrimaryKey: true},
            {Name: "name", Type: sb.COLUMN_TYPE_TEXT},
            {Name: "status", Type: sb.COLUMN_TYPE_TEXT},
        },
    })

    if err != nil {
        t.Fatalf("NewStore: %v", err)
    }

    if err := store.InsertMany([]map[string]any{{"id": 1, "name": "Alice", "status": "ok"}, {"id": 2, "name": "Bob", "status": "ok"}, {"id": 3, "name": "Charlie", "status": "ok"}}); err != nil {
        t.Fatalf("InsertMany: %v", err)
    }

    if err := store.DeleteWhereEquals(map[string]any{"name": "Bob"}); err != nil {
        t.Fatalf("DeleteWhereEquals: %v", err)
    }

    // Expect 2 rows left, and none with name Bob
    count, err := store.Count(SearchQuery{})
    if err != nil {
        t.Fatalf("Count: %v", err)
    }
    if count != 2 {
        t.Fatalf("Count after delete: expected 2 got %d", count)
    }

    res, err := store.Search(SearchQuery{Where: goqu.Ex{"name": "Bob"}})
    if err != nil {
        t.Fatalf("Search: %v", err)
    }
    if len(res) != 0 {
        t.Fatalf("Expected no rows for Bob, got %d", len(res))
    }
}

func Test_Store_UpsertWhereEquals_InsertThenUpdate(t *testing.T) {
    db := initDB("test_store_upsert_where_equals.db")

    store, err := NewStore(NewStoreOptions{
        TableName:          "test_index_upsert_filters",
        DB:                 db,
        AutomigrateEnabled: true,
        Columns: []sb.Column{
            {Name: "id", Type: sb.COLUMN_TYPE_INTEGER, PrimaryKey: true},
            {Name: "name", Type: sb.COLUMN_TYPE_TEXT},
            {Name: "status", Type: sb.COLUMN_TYPE_TEXT},
        },
    })

    if err != nil {
        t.Fatalf("NewStore: %v", err)
    }

    // Insert via UpsertWhereEquals when no row exists
    if err := store.UpsertWhereEquals(map[string]any{"id": 1}, map[string]any{"id": 1, "name": "Alice", "status": "pending"}); err != nil {
        t.Fatalf("UpsertWhereEquals insert: %v", err)
    }

    res, err := store.Search(SearchQuery{Where: goqu.Ex{"id": 1}})
    if err != nil || len(res) != 1 {
        t.Fatalf("Search after insert: err=%v len=%d", err, len(res))
    }
    if cast.ToString(res[0]["name"]) != "Alice" {
        t.Fatalf("Expected name Alice, got %v", res[0]["name"])
    }

    // Update via UpsertWhereEquals (non-filter columns only)
    if err := store.UpsertWhereEquals(map[string]any{"id": 1}, map[string]any{"name": "Alice2", "status": "active"}); err != nil {
        t.Fatalf("UpsertWhereEquals update: %v", err)
    }

    res, err = store.Search(SearchQuery{Where: goqu.Ex{"id": 1}})
    if err != nil || len(res) != 1 {
        t.Fatalf("Search after update: err=%v len=%d", err, len(res))
    }
    if cast.ToString(res[0]["name"]) != "Alice2" {
        t.Fatalf("Expected name Alice2, got %v", res[0]["name"])
    }
    if cast.ToString(res[0]["status"]) != "active" {
        t.Fatalf("Expected status active, got %v", res[0]["status"])
    }
}

func Test_Store_Upsert_Wrapper(t *testing.T) {
    db := initDB("test_store_upsert_wrapper.db")

    store, err := NewStore(NewStoreOptions{
        TableName:          "test_index_upsert_wrapper",
        DB:                 db,
        AutomigrateEnabled: true,
        Columns: []sb.Column{
            {Name: "id", Type: sb.COLUMN_TYPE_INTEGER, PrimaryKey: true},
            {Name: "name", Type: sb.COLUMN_TYPE_TEXT},
        },
    })

    if err != nil {
        t.Fatalf("NewStore: %v", err)
    }

    // Insert via Upsert with conflictColumns ["id"]
    if err := store.Upsert(map[string]any{"id": 1, "name": "Bob"}, []string{"id"}); err != nil {
        t.Fatalf("Upsert insert: %v", err)
    }

    // Update existing row via Upsert
    if err := store.Upsert(map[string]any{"id": 1, "name": "Bobby"}, []string{"id"}); err != nil {
        t.Fatalf("Upsert update: %v", err)
    }

    res, err := store.Search(SearchQuery{Where: goqu.Ex{"id": 1}})
    if err != nil || len(res) != 1 {
        t.Fatalf("Search after upsert: err=%v len=%d", err, len(res))
    }
    if cast.ToString(res[0]["name"]) != "Bobby" {
        t.Fatalf("Expected name Bobby, got %v", res[0]["name"])
    }

    // Missing conflict column value should return error
    if err := store.Upsert(map[string]any{"name": "NoID"}, []string{"id"}); err == nil {
        t.Fatalf("Expected error when missing conflict column value")
    }
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
	db := initDB("test_store_insert.db")

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
	db := initDB("test_store_insertmany.db")

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
	db := initDB("test_store_search_selectall.db")

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
	db := initDB("test_store_search_where.db")

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
	db := initDB("test_store_search_whereexpr.db")

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
	db := initDB("test_store_count.db")

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
