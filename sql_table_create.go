package indexstore

import "github.com/dracory/sb"

func (store *Store) sqlTableCreate() string {
	columns := store.columns

	sql := sb.NewBuilder(sb.DatabaseDriverName(store.db)).
		Table(store.tableName)

	for _, column := range columns {
		sql = sql.Column(column)
	}

	sqlStr, err := sql.CreateIfNotExists()
	if err != nil {
		return ""
	}

	return sqlStr
}

func (store *Store) sqlTableDrop() (string, error) {
	sql, err := sb.NewBuilder(sb.DatabaseDriverName(store.db)).
		Table(store.tableName).
		DropIfExists()
	if err != nil {
		return "", err
	}
	return sql, nil
}
