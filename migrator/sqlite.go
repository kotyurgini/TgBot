package migrator

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Sqlite struct {
	storagePath    string
	migrationsPath string
	lg             *slog.Logger
}

func NewSqliteMigrator(storagePath string, lg *slog.Logger) (Migrator, error) {
	var err error
	s := &Sqlite{storagePath: storagePath, lg: lg}
	s.migrationsPath, err = getMigrationsPath("sqlite")
	return s, err
}

func (s *Sqlite) MustMigrate() {
	if s.storagePath == "" {
		panic("storage-path is empty")
	}
	if s.migrationsPath == "" {
		panic("migrations-path is empty")
	}

	m, err := migrate.New(
		"file://"+s.migrationsPath,
		fmt.Sprintf("sqlite://%s?x-migrations-table=%s", s.storagePath, migrationsTableName),
	)
	if err != nil {
		panic(err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			s.lg.Info("No migrations to apply")
			return
		}
		panic(err)
	}

	s.lg.Info("Migrations applied")
}
