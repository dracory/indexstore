package indexstore

import (
	"database/sql"
	"testing"

	"github.com/spf13/cast"
	_ "modernc.org/sqlite"
)

func initDB(_ string) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	return db
}

func Test_Store_DeleteWhereEquals(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_delete",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []ColumnDefinition{
			{Name: "id", Type: "string", PrimaryKey: true},
			{Name: "data", Type: "text"},
		},
	})

	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	if err := store.InsertMany([]map[string]any{{"id": "1", "data": "Alice"}, {"id": "2", "data": "Bob"}, {"id": "3", "data": "Charlie"}}); err != nil {
		t.Fatalf("InsertMany: %v", err)
	}

	if err := store.DeleteWhereEquals(map[string]any{"data": "Bob"}); err != nil {
		t.Fatalf("DeleteWhereEquals: %v", err)
	}

	// Expect 2 rows left, and none with data Bob
	count, err := store.Count(SearchQuery{})
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 2 {
		t.Fatalf("Count after delete: expected 2 got %d", count)
	}

	res, err := store.Search(SearchQuery{Where: map[string]any{"data": "Bob"}})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("Expected no rows for Bob, got %d", len(res))
	}
}

func Test_Store_UpsertWhereEquals_InsertThenUpdate(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_upsert_filters",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []ColumnDefinition{
			{Name: "id", Type: "string", PrimaryKey: true},
			{Name: "data", Type: "text"},
		},
	})

	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	// Insert via UpsertWhereEquals when no row exists
	if err := store.UpsertWhereEquals(map[string]any{"id": "1"}, map[string]any{"id": "1", "data": "Alice"}); err != nil {
		t.Fatalf("UpsertWhereEquals insert: %v", err)
	}

	res, err := store.Search(SearchQuery{Where: map[string]any{"id": "1"}})
	if err != nil || len(res) != 1 {
		t.Fatalf("Search after insert: err=%v len=%d", err, len(res))
	}
	if cast.ToString(res[0]["data"]) != "Alice" {
		t.Fatalf("Expected data Alice, got %v", res[0]["data"])
	}

	// Update via UpsertWhereEquals (non-filter columns only)
	if err := store.UpsertWhereEquals(map[string]any{"id": "1"}, map[string]any{"data": "Alice2"}); err != nil {
		t.Fatalf("UpsertWhereEquals update: %v", err)
	}

	res, err = store.Search(SearchQuery{Where: map[string]any{"id": "1"}})
	if err != nil || len(res) != 1 {
		t.Fatalf("Search after update: err=%v len=%d", err, len(res))
	}
	if cast.ToString(res[0]["data"]) != "Alice2" {
		t.Fatalf("Expected data Alice2, got %v", res[0]["data"])
	}
}

func Test_Store_Upsert_Wrapper(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_upsert_wrapper",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []ColumnDefinition{
			{Name: "id", Type: "string", PrimaryKey: true},
			{Name: "data", Type: "text"},
		},
	})

	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	// Insert via Upsert with conflictColumns ["id"]
	if err := store.Upsert(map[string]any{"id": "1", "data": "Bob"}, []string{"id"}); err != nil {
		t.Fatalf("Upsert insert: %v", err)
	}

	// Update existing row via Upsert
	if err := store.Upsert(map[string]any{"id": "1", "data": "Bobby"}, []string{"id"}); err != nil {
		t.Fatalf("Upsert update: %v", err)
	}

	res, err := store.Search(SearchQuery{Where: map[string]any{"id": "1"}})
	if err != nil || len(res) != 1 {
		t.Fatalf("Search after upsert: err=%v len=%d", err, len(res))
	}
	if cast.ToString(res[0]["data"]) != "Bobby" {
		t.Fatalf("Expected data Bobby, got %v", res[0]["data"])
	}

	// Missing conflict column value should return error
	if err := store.Upsert(map[string]any{"data": "NoID"}, []string{"id"}); err == nil {
		t.Fatalf("Expected error when missing conflict column value")
	}
}

func Test_Store_WithAutoMigrate(t *testing.T) {
	db := initDB("test_store_with_automigrate.db")

	storeAutomigrateFalse, errAutomigrateFalse := NewStore(NewStoreOptions{
		TableName:          "test_index_with_automigrate_false",
		DB:                 db,
		AutomigrateEnabled: false,
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
	})

	if errAutomigrateTrue != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", errAutomigrateTrue.Error())
	}

	if storeAutomigrateTrue.automigrateEnabled != true {
		t.Fatalf("automigrateEnabled: Expected [true] received [%v]", storeAutomigrateTrue.automigrateEnabled)
	}
}

func Test_Store_Insert(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_search",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []ColumnDefinition{
			{Name: "id", Type: "string", PrimaryKey: true},
		},
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.Insert(map[string]any{
		"id": "1",
	})

	if err != nil {
		t.Fatalf("Insert: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.Insert(map[string]any{
		"id": "2",
	})

	if err != nil {
		t.Fatalf("Insert: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.Insert(map[string]any{
		"id": "3",
	})

	if err != nil {
		t.Fatalf("Insert: Expected [err] to be nill received [%v]", err.Error())
	}
}

func Test_Store_InsertMany(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_search",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []ColumnDefinition{
			{Name: "id", Type: "string", PrimaryKey: true},
		},
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.InsertMany([]map[string]any{
		{
			"id": "1",
		},
		{
			"id": "2",
		},
		{
			"id": "3",
		},
	})

	if err != nil {
		t.Fatalf("InsertMany: Expected [err] to be nill received [%v]", err.Error())
	}
}

func Test_Store_Search_SelectAll(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_search",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []ColumnDefinition{
			{Name: "id", Type: "string", PrimaryKey: true},
			{Name: "data", Type: "text"},
		},
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.InsertMany([]map[string]any{
		{"id": "1", "data": "Alice"},
		{"id": "2", "data": "Bob"},
		{"id": "3", "data": "Charlie"},
	})

	if err != nil {
		t.Fatalf("InsertMany: Expected [err] to be nill received [%v]", err.Error())
	}

	res, err := store.Search(SearchQuery{})

	if err != nil {
		t.Fatalf("Search: Expected [err] to be nill received [%v]", err.Error())
	}

	if len(res) != 3 {
		t.Fatalf("Search: Expected [3] results received [%d]", len(res))
	}
}

func Test_Store_Search_WithWhere(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_search",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []ColumnDefinition{
			{Name: "id", Type: "string", PrimaryKey: true},
			{Name: "data", Type: "text"},
		},
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.InsertMany([]map[string]any{
		{"id": "1", "data": "Alice"},
		{"id": "2", "data": "Bob"},
		{"id": "3", "data": "Charlie"},
	})

	if err != nil {
		t.Fatalf("InsertMany: Expected [err] to be nill received [%v]", err.Error())
	}

	res, err := store.Search(SearchQuery{
		Where: map[string]any{"data": "Bob"},
	})

	if err != nil {
		t.Fatalf("Search: Expected [err] to be nill received [%v]", err.Error())
	}

	if len(res) != 1 {
		t.Fatalf("Search: Expected [1] results received [%d]", len(res))
	}

	if cast.ToString(res[0]["data"]) != "Bob" {
		t.Fatalf("Search: Expected [Bob] received [%s]", cast.ToString(res[0]["data"]))
	}
}

func Test_Store_Search_WithOrderBy(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_search",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []ColumnDefinition{
			{Name: "id", Type: "string", PrimaryKey: true},
			{Name: "data", Type: "text"},
		},
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.InsertMany([]map[string]any{
		{"id": "3", "data": "Charlie"},
		{"id": "1", "data": "Alice"},
		{"id": "2", "data": "Bob"},
	})

	if err != nil {
		t.Fatalf("InsertMany: Expected [err] to be nill received [%v]", err.Error())
	}

	res, err := store.Search(SearchQuery{
		OrderBy:   "id",
		SortOrder: "asc",
	})

	if err != nil {
		t.Fatalf("Search: Expected [err] to be nill received [%v]", err.Error())
	}

	if len(res) != 3 {
		t.Fatalf("Search: Expected [3] results received [%d]", len(res))
	}

	if cast.ToString(res[0]["id"]) != "1" {
		t.Fatalf("Search: Expected [1] received [%s]", cast.ToString(res[0]["id"]))
	}
}

func Test_Store_Search_WithLimit(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_search",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []ColumnDefinition{
			{Name: "id", Type: "string", PrimaryKey: true},
			{Name: "data", Type: "text"},
		},
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.InsertMany([]map[string]any{
		{"id": "1", "data": "Alice"},
		{"id": "2", "data": "Bob"},
		{"id": "3", "data": "Charlie"},
	})

	if err != nil {
		t.Fatalf("InsertMany: Expected [err] to be nill received [%v]", err.Error())
	}

	res, err := store.Search(SearchQuery{
		Limit: 2,
	})

	if err != nil {
		t.Fatalf("Search: Expected [err] to be nill received [%v]", err.Error())
	}

	if len(res) != 2 {
		t.Fatalf("Search: Expected [2] results received [%d]", len(res))
	}
}

func Test_Store_Search_WithOffset(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_search",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []ColumnDefinition{
			{Name: "id", Type: "string", PrimaryKey: true},
			{Name: "data", Type: "text"},
		},
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.InsertMany([]map[string]any{
		{"id": "1", "data": "Alice"},
		{"id": "2", "data": "Bob"},
		{"id": "3", "data": "Charlie"},
	})

	if err != nil {
		t.Fatalf("InsertMany: Expected [err] to be nill received [%v]", err.Error())
	}

	res, err := store.Search(SearchQuery{
		Offset: 1,
	})

	if err != nil {
		t.Fatalf("Search: Expected [err] to be nill received [%v]", err.Error())
	}

	if len(res) != 2 {
		t.Fatalf("Search: Expected [2] results received [%d]", len(res))
	}
}

func Test_Store_Count(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_search",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []ColumnDefinition{
			{Name: "id", Type: "string", PrimaryKey: true},
			{Name: "data", Type: "text"},
		},
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.InsertMany([]map[string]any{
		{"id": "1", "data": "Alice"},
		{"id": "2", "data": "Bob"},
		{"id": "3", "data": "Charlie"},
	})

	if err != nil {
		t.Fatalf("InsertMany: Expected [err] to be nill received [%v]", err.Error())
	}

	count, err := store.Count(SearchQuery{})

	if err != nil {
		t.Fatalf("Count: Expected [err] to be nill received [%v]", err.Error())
	}

	if count != 3 {
		t.Fatalf("Count: Expected [3] received [%d]", count)
	}
}

func Test_Store_Truncate(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_search",
		DB:                 db,
		AutomigrateEnabled: true,
		Columns: []ColumnDefinition{
			{Name: "id", Type: "string", PrimaryKey: true},
			{Name: "data", Type: "text"},
		},
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.InsertMany([]map[string]any{
		{"id": "1", "data": "Alice"},
		{"id": "2", "data": "Bob"},
		{"id": "3", "data": "Charlie"},
	})

	if err != nil {
		t.Fatalf("InsertMany: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.Truncate()

	if err != nil {
		t.Fatalf("Truncate: Expected [err] to be nill received [%v]", err.Error())
	}

	count, err := store.Count(SearchQuery{})

	if err != nil {
		t.Fatalf("Count: Expected [err] to be nill received [%v]", err.Error())
	}

	if count != 0 {
		t.Fatalf("Count: Expected [0] received [%d]", count)
	}
}

func Test_Store_Drop(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_search",
		DB:                 db,
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.Drop()

	if err != nil {
		t.Fatalf("Drop: Expected [err] to be nill received [%v]", err.Error())
	}
}

func Test_Store_MigrateUp(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_migrateup",
		DB:                 db,
		AutomigrateEnabled: false,
	})

	if err != nil {
		t.Fatalf("NewStore: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.MigrateUp(nil)

	if err != nil {
		t.Fatalf("MigrateUp: Expected [err] to be nill received [%v]", err.Error())
	}
}

func Test_Store_MigrateDown(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_migratedown",
		DB:                 db,
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("NewStore: Expected [err] to be nill received [%v]", err.Error())
	}

	err = store.MigrateDown(nil)

	if err != nil {
		t.Fatalf("MigrateDown: Expected [err] to be nill received [%v]", err.Error())
	}
}

func Test_Store_Debug(t *testing.T) {
	db := initDB("")

	store, err := NewStore(NewStoreOptions{
		TableName:          "test_index_debug",
		DB:                 db,
		AutomigrateEnabled: true,
		DebugEnabled:       false,
	})

	if err != nil {
		t.Fatalf("NewStore: Expected [err] to be nill received [%v]", err.Error())
	}

	store.Debug(true)

	if store.debugEnabled != true {
		t.Fatalf("Debug: Expected [true] received [%v]", store.debugEnabled)
	}

	store.Debug(false)

	if store.debugEnabled != false {
		t.Fatalf("Debug: Expected [false] received [%v]", store.debugEnabled)
	}
}
