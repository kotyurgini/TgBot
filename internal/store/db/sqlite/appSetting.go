package sqlite

import (
	"context"
	"tgbot/common"
	"tgbot/internal/store"
)

func (d *DB) AppSettingGet(ctx context.Context) (*store.AppSetting, error) {
	q := `SELECT * FROM appSettings LIMIT 1`
	entity := &store.AppSetting{}
	if err := d.db.QueryRowContext(ctx, q).Scan(&entity.ID, &entity.Maintenance); err != nil {
		return nil, common.WrapErrors("AppSettingGet()", store.ErrDBQueryError, err)
	}
	return entity, nil
}

func (d *DB) AppSettingUpdate(ctx context.Context, entity *store.AppSetting) (*store.AppSetting, error) {
	q := `UPDATE appSettings SET maintenance = ? WHERE id = ?`
	if _, err := d.db.ExecContext(ctx, q, entity.Maintenance, entity.ID); err != nil {
		return nil, common.WrapErrors("AppSettingUpdate()", store.ErrDBQueryError, err)
	}
	return entity, nil
}
