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
