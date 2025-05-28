package sqlite

import (
	"context"
	"strings"
	"tgbot/common"
	"tgbot/internal/store"
)

func (d *DB) ActiveDialogUpsert(ctx context.Context, entity *store.ActiveDialog) (*store.ActiveDialog, error) {
	q := `INSERT INTO activeDialogs (userId, dialogId)
			VALUES (?, ?)
			ON CONFLICT(userId) DO UPDATE SET
				dialogId = excluded.dialogId;`

	_, err := d.db.ExecContext(ctx, q, entity.UserID, entity.DialogID)
	if err != nil {
		return nil, common.WrapErrors("ActiveDialogUpsert()", store.ErrDBQueryError, err)
	}

	return entity, nil
}

func (d *DB) ActiveDialogList(ctx context.Context, filter *store.ActiveDialogFilter) ([]*store.ActiveDialog, error) {
	method := "ActiveDialogList()"
	where, args := []string{"1 = 1"}, []any{}

	if filter.UserID != nil {
		where, args = append(where, "userId = ?"), append(args, filter.UserID)
	}
	if filter.DialogID != nil {
		where, args = append(where, "dialogId = ?"), append(args, filter.DialogID)
	}

	q := `
		SELECT *		
		FROM activeDialogs
		WHERE ` + strings.Join(where, " AND ")

	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, common.WrapErrors(method, store.ErrDBQueryError, err)
	}
	defer closeRows(rows)

	list := make([]*store.ActiveDialog, 0)
	for rows.Next() {
		var entity store.ActiveDialog
		if err := rows.Scan(
			&entity.UserID,
			&entity.DialogID,
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

func (d *DB) ActiveDialogDelete(ctx context.Context, filter *store.ActiveDialogFilter) error {
	method := "ActiveDialogList()"
	where, args := []string{}, []any{}

	if filter.UserID != nil {
		where, args = append(where, "userId = ?"), append(args, filter.UserID)
	}
	if filter.DialogID != nil {
		where, args = append(where, "dialogId = ?"), append(args, filter.DialogID)
	}

	if len(where) == 0 {
		return common.WrapErrors(method, store.ErrDBNoFilterProvided)
	}

	q := `
		DELETE	
		FROM activeDialogs
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
