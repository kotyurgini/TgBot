package sqlite

import (
	"database/sql"
	"tgbot/internal/store"

	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
}

func NewDB(path string) (store.Driver, error) {
	sqliteDB, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)")
	// if err != nil {
	// 	return nil, errors.Wrapf(err, "failed to open db with dsn: %s", profile.DSN)
	// }
	if err != nil {
		return nil, err
	}

	if err := sqliteDB.Ping(); err != nil {
		return nil, err
	}

	driver := DB{db: sqliteDB}

	return &driver, nil
}

func (d *DB) GetDB() *sql.DB {
	return d.db
}

func (d *DB) Close() error {
	return d.db.Close()
}
