package sqlite

import (
	"context"
	"fmt"
	"strings"
	"tgbot/common"
	"tgbot/internal/store"
	"time"
)

func (d *DB) DialogCreate(ctx context.Context, entity *store.Dialog) (*store.Dialog, error) {
	fields := []string{"userId", "title", "created"}
	args := []any{entity.UserID, entity.Title, entity.Created}

	q := "INSERT INTO dialogs (" + strings.Join(fields, ", ") + ") VALUES (" + placeholders(len(fields)) + ") RETURNING id"

	if err := d.db.QueryRowContext(ctx, q, args...).Scan(
		&entity.ID,
	); err != nil {
		return nil, common.WrapErrors("DialogCreate()", store.ErrDBQueryError, err)
	}

	return entity, nil
}

func (d *DB) DialogList(ctx context.Context, filter *store.DialogFilter) (int, []*store.Dialog, error) {
	method := "DialogList()"
	where, args := []string{}, []any{}

	if filter.ID != nil {
		where, args = append(where, "id = ?"), append(args, filter.ID)
	}
	if filter.UserID != nil {
		where, args = append(where, "userId = ?"), append(args, filter.UserID)
	}
	if filter.Title != nil {
		where, args = append(where, "title = ?"), append(args, filter.Title)
	}
	if filter.Created != nil {
		where, args = append(where, "created = ?"), append(args, filter.Created)
	}

	qCount := `
		SELECT COUNT(*)
		FROM dialogs
		WHERE ` + strings.Join(where, " AND ")
	count := 0
	if err := d.db.QueryRowContext(ctx, qCount, args...).Scan(&count); err != nil {
		return 0, nil, common.WrapErrors(method, store.ErrDBQueryError, err)
	}

	q := `
		SELECT *	
		FROM dialogs
		WHERE ` + strings.Join(where, " AND ") +
		`ORDER BY id `
	if filter.Limit > 0 {
		q += `LIMIT ? `
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		q += `OFFSET ? `
		args = append(args, filter.Offset)
	}

	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return 0, nil, common.WrapErrors(method, store.ErrDBQueryError, err)
	}
	defer closeRows(rows)

	list := make([]*store.Dialog, 0)
	for rows.Next() {
		var entity store.Dialog
		var created string
		if err := rows.Scan(
			&entity.ID,
			&entity.UserID,
			&entity.Title,
			&created,
		); err != nil {
			return 0, nil, common.WrapErrors(method, store.ErrDBScanRowError, err)
		}
		entity.Created, err = time.Parse(dateLayout(), created)
		if err != nil {
			return 0, nil, fmt.Errorf("%s: failed to parse created: %w", method, err)
		}
		list = append(list, &entity)
	}

	if err := rows.Err(); err != nil {
		return 0, nil, common.WrapErrors(method, store.ErrDBRowError, err)
	}

	return count, list, nil
}

func (d *DB) DialogUpdate(ctx context.Context, entity *store.Dialog) (*store.Dialog, error) {
	q := `UPDATE dialogs
			SET
				userId = ?,
				title = ?,
				created = ?
			WHERE
				id = ?;`

	_, err := d.db.ExecContext(ctx, q,
		entity.UserID,
		entity.Title,
		entity.Created,
		entity.ID)
	if err != nil {
		return nil, common.WrapErrors("DialogUpdate()", store.ErrDBQueryError, err)
	}
	return entity, nil
}

func (d *DB) DialogDelete(ctx context.Context, filter *store.DialogFilter) error {
	method := "DialogDelete()"
	where, args := []string{}, []any{}

	if filter.ID != nil {
		where, args = append(where, "id = ?"), append(args, filter.ID)
	}
	if filter.UserID != nil {
		where, args = append(where, "userId = ?"), append(args, filter.UserID)
	}
	if filter.Title != nil {
		where, args = append(where, "title = ?"), append(args, filter.Title)
	}
	if filter.Created != nil {
		where, args = append(where, "created = ?"), append(args, filter.Created)
	}

	if len(where) == 0 {
		return common.WrapErrors(method, store.ErrDBNoFilterProvided)
	}

	q := `
		DELETE	
		FROM dialogs
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
