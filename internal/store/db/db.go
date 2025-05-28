package db

import (
	"errors"
	"fmt"
	"tgbot/internal/config"
	"tgbot/internal/store"
	"tgbot/internal/store/db/sqlite"
)

func NewDBDriver(cfg *config.Config) (store.Driver, error) {
	var driver store.Driver
	var err error

	switch cfg.DbDriver {
	case "sqlite":
		driver, err = sqlite.NewDB(cfg.StoragePath)
	// case "postgres":
	// 	driver, err = postgres.NewDB(profile)
	default:
		return nil, errors.New("unknown db driver")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create db driver: %w", err)
	}
	return driver, nil
}
