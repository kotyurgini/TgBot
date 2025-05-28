package sqlite

import (
	"context"
	"strings"
	"tgbot/common"
	"tgbot/internal/store"
)

func (d *DB) TariffLimitCreate(ctx context.Context, entity *store.TariffLimit) (*store.TariffLimit, error) {
	fields := []string{"tariffId", "aiModelId", "count"}
	args := []any{entity.TariffID, entity.AIModelID, entity.Count}

	q := "INSERT INTO tariffLimits (" + strings.Join(fields, ", ") + ") VALUES (" + placeholders(len(fields)) + ") RETURNING id"

	if err := d.db.QueryRowContext(ctx, q, args...).Scan(
		&entity.ID,
	); err != nil {
		return nil, common.WrapErrors("TariffLimitCreate()", store.ErrDBQueryError, err)
	}

	return entity, nil
}

func (d *DB) TariffLimitList(ctx context.Context, filter *store.TariffLimitFilter) ([]*store.TariffLimit, error) {
	method := "TariffLimitList()"
	where, args := []string{"1 = 1"}, []any{}

	if filter.ID != nil {
		where, args = append(where, "id = ?"), append(args, filter.ID)
	}
	if filter.TariffID != nil {
		where, args = append(where, "tariffId = ?"), append(args, filter.TariffID)
	}
	if filter.AIModelID != nil {
		where, args = append(where, "aiModelId = ?"), append(args, filter.AIModelID)
	}
	if filter.Count != nil {
		where, args = append(where, "count = ?"), append(args, filter.Count)
	}

	q := `
		SELECT *	
		FROM tariffLimits
		WHERE ` + strings.Join(where, " AND ")

	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, common.WrapErrors(method, store.ErrDBQueryError, err)
	}
	defer closeRows(rows)

	list := make([]*store.TariffLimit, 0)
	for rows.Next() {
		var entity store.TariffLimit
		if err := rows.Scan(
			&entity.ID,
			&entity.TariffID,
			&entity.AIModelID,
			&entity.Count,
		); err != nil {
			return nil, common.WrapErrors(method, store.ErrDBScanRowError, err)
		}
		list = append(list, &entity)
	}

	if err := rows.Err(); err != nil {
		return nil, common.WrapErrors(method, store.ErrDBRowError, err)
	}

	return list, nil
}

func (d *DB) TariffLimitUpdate(ctx context.Context, entity *store.TariffLimit) (*store.TariffLimit, error) {
	q := `UPDATE tariffLimits
			SET
				tariffId = ?,
				aiModelId = ?,
				count = ?
			WHERE
				id = ?;`

	_, err := d.db.ExecContext(ctx, q,
		entity.TariffID,
		entity.AIModelID,
		entity.Count,
		entity.ID)
	if err != nil {
		return nil, common.WrapErrors("TariffLimitUpdate()", store.ErrDBQueryError, err)
	}
	return entity, nil
}

func (d *DB) TariffLimitDelete(ctx context.Context, filter *store.TariffLimitFilter) error {
	method := "TariffLimitDelete()"
	where, args := []string{}, []any{}

	if filter.ID != nil {
		where, args = append(where, "id = ?"), append(args, filter.ID)
	}
	if filter.TariffID != nil {
		where, args = append(where, "tariffId = ?"), append(args, filter.TariffID)
	}
	if filter.AIModelID != nil {
		where, args = append(where, "aiModelId = ?"), append(args, filter.AIModelID)
	}
	if filter.Count != nil {
		where, args = append(where, "count = ?"), append(args, filter.Count)
	}

	if len(where) == 0 {
		return common.WrapErrors(method, store.ErrDBNoFilterProvided)
	}

	q := `
		DELETE	
		FROM tariffLimits
		WHERE ` + strings.Join(where, " AND ")

	result, err := d.db.ExecContext(ctx, q, args...)
	if err != nil {
		return common.WrapErrors(method, store.ErrDBQueryError, err)
	}
	if _, err := result.RowsAffected(); err != nil {
		return common.WrapErrors(method, store.ErrDBNoRowsAffected)
	}

	return nil
}
