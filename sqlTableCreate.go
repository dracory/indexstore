package indexstore

import "github.com/gouniverse/sb"

func (store *Store) sqlTableCreate() string {
	columns := store.columns

	sql := sb.NewBuilder(sb.DatabaseDriverName(store.db)).
		Table(store.tableName)

	for _, column := range columns {
		sql = sql.Column(column)
	}

	sqlStr := sql.CreateIfNotExists()

	return sqlStr
}
