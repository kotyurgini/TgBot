package migrator

import "fmt"

const migrationsTableName = "_migrations"

type Migrator interface {
	MustMigrate()
}

func getMigrationsPath(dbDriver string) (string, error) {
	switch dbDriver {
	case "sqlite":
		return "./migrations/sqlite", nil
	default:
		return "", fmt.Errorf("unsupported db driver: %s", dbDriver)
	}
}
