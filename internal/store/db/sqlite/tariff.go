package sqlite

import (
	"context"
	"strings"
	"tgbot/common"
	"tgbot/internal/store"
)

func (d *DB) TariffCreate(ctx context.Context, entity *store.Tariff) (*store.Tariff, error) {
	fields := []string{"title", "rubPrice", "usdPrice", "available"}
	args := []any{entity.Title, entity.RubPrice, entity.UsdPrice, entity.Available}

	q := "INSERT INTO tariffs (" + strings.Join(fields, ", ") + ") VALUES (" + placeholders(len(fields)) + ") RETURNING id"

	if err := d.db.QueryRowContext(ctx, q, args...).Scan(
		&entity.ID,
	); err != nil {
		return nil, common.WrapErrors("TariffCreate()", store.ErrDBQueryError, err)
	}

	return entity, nil
}

func (d *DB) TariffList(ctx context.Context, filter *store.TariffFilter) ([]*store.Tariff, error) {
	method := "TariffList()"
	where, args := []string{"1 = 1"}, []any{}

	if filter.ID != nil {
		where, args = append(where, "id = ?"), append(args, filter.ID)
	}
	if filter.Title != nil {
		where, args = append(where, "title = ?"), append(args, filter.Title)
	}
	if filter.RubPrice != nil {
		where, args = append(where, "rubPrice = ?"), append(args, filter.RubPrice)
	}
	if filter.UsdPrice != nil {
		where, args = append(where, "usdPrice = ?"), append(args, filter.UsdPrice)
	}
	if filter.Available != nil {
		where, args = append(where, "available = ?"), append(args, filter.Available)
	}

	q := `
		SELECT *	
		FROM tariffs
		WHERE ` + strings.Join(where, " AND ")

	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, common.WrapErrors(method, store.ErrDBQueryError, err)
	}
	defer closeRows(rows)

	list := make([]*store.Tariff, 0)
	for rows.Next() {
		var entity store.Tariff
		if err := rows.Scan(
			&entity.ID,
			&entity.Title,
			&entity.RubPrice,
			&entity.UsdPrice,
			&entity.Available,
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

func (d *DB) TariffUpdate(ctx context.Context, entity *store.Tariff) (*store.Tariff, error) {
	q := `UPDATE tariffs
			SET
				title = ?,
				enTitle = ?,
				rubPrice = ?,
				usdPrice = ?,
				available = ?
			WHERE
				id = ?;`

	_, err := d.db.ExecContext(ctx, q,
		entity.Title,
		entity.RubPrice,
		entity.UsdPrice,
		entity.Available,
		entity.ID)
	if err != nil {
		return nil, common.WrapErrors("TariffUpdate()", store.ErrDBQueryError, err)
	}
	return entity, nil
}

func (d *DB) TariffDelete(ctx context.Context, filter *store.TariffFilter) error {
	method := "TariffDelete()"
	where, args := []string{}, []any{}

	if filter.ID != nil {
		where, args = append(where, "id = ?"), append(args, filter.ID)
	}
	if filter.Title != nil {
		where, args = append(where, "title = ?"), append(args, filter.Title)
	}
	if filter.RubPrice != nil {
		where, args = append(where, "rubPrice = ?"), append(args, filter.RubPrice)
	}
	if filter.UsdPrice != nil {
		where, args = append(where, "usdPrice = ?"), append(args, filter.UsdPrice)
	}
	if filter.Available != nil {
		where, args = append(where, "available = ?"), append(args, filter.Available)
	}

	if len(where) == 0 {
		return common.WrapErrors(method, store.ErrDBNoFilterProvided)
	}

	q := `
		DELETE	
		FROM tariffs
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
