package sqlite

import (
	"context"
	"fmt"
	"strings"
	"tgbot/common"
	"tgbot/internal/store"
	"time"
)

func (d *DB) UserUsageCreate(ctx context.Context, entity *store.UserUsage) (*store.UserUsage, error) {
	fields := []string{"userId", "aiModelId", "count"}
	args := []any{entity.UserID, entity.AIModelID, entity.Count}

	q := "INSERT INTO usersUsage (" + strings.Join(fields, ", ") + ") VALUES (" + placeholders(len(fields)) + ")"

	if _, err := d.db.ExecContext(ctx, q, args...); err != nil {
		return nil, common.WrapErrors("UserUsageCreate()", store.ErrDBQueryError, err)
	}

	return entity, nil
}

func (d *DB) UserUsageList(ctx context.Context, filter *store.UserUsageFilter) ([]*store.UserUsage, error) {
	method := "UserUsageList()"
	where, args := []string{"1 = 1"}, []any{}

	if filter.UserID != nil {
		where, args = append(where, "userId = ?"), append(args, filter.UserID)
	}
	if filter.AIModelID != nil {
		where, args = append(where, "aiModelId = ?"), append(args, filter.AIModelID)
	}
	if filter.Count != nil {
		where, args = append(where, "count = ?"), append(args, filter.Count)
	}
	if filter.LastActivity != nil {
		where, args = append(where, "lastActivity = ?"), append(args, filter.LastActivity)
	}

	q := `
		SELECT *	
		FROM usersUsage
		WHERE ` + strings.Join(where, " AND ")

	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, common.WrapErrors(method, store.ErrDBQueryError, err)
	}
	defer closeRows(rows)

	list := make([]*store.UserUsage, 0)
	for rows.Next() {
		var entity store.UserUsage
		var lastActivity string
		if err := rows.Scan(
			&entity.UserID,
			&entity.AIModelID,
			&entity.Count,
			&lastActivity,
		); err != nil {
			return nil, common.WrapErrors(method, store.ErrDBScanRowError, err)
		}
		entity.LastActivity, err = time.Parse(dateLayout(), lastActivity)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to parse lastActivity: %w", method, err)
		}
		list = append(list, &entity)
	}

	if err := rows.Err(); err != nil {
		return nil, common.WrapErrors(method, store.ErrDBRowError, err)
	}

	return list, nil
}

func (d *DB) UserUsageUpdate(ctx context.Context, entity *store.UserUsage) (*store.UserUsage, error) {
	q := `UPDATE usersUsage
			SET
				count = ?,
				lastActivity = ?
			WHERE
				userId = ? AND aiModelId = ?;`

	_, err := d.db.ExecContext(ctx, q,
		entity.Count,
		entity.LastActivity,
		entity.UserID,
		entity.AIModelID)
	if err != nil {
		return nil, common.WrapErrors("UserUsageUpdate()", store.ErrDBQueryError, err)
	}
	return entity, nil
}

func (d *DB) UserUsageDelete(ctx context.Context, filter *store.UserUsageFilter) error {
	method := "UserUsageDelete()"
	where, args := []string{}, []any{}

	if filter.UserID != nil {
		where, args = append(where, "userId = ?"), append(args, filter.UserID)
	}
	if filter.AIModelID != nil {
		where, args = append(where, "aiModelId = ?"), append(args, filter.AIModelID)
	}
	if filter.Count != nil {
		where, args = append(where, "count = ?"), append(args, filter.Count)
	}
	if filter.LastActivity != nil {
		where, args = append(where, "lastActivity = ?"), append(args, filter.LastActivity)
	}

	if len(where) == 0 {
		return common.WrapErrors(method, store.ErrDBNoFilterProvided)
	}

	q := `
		DELETE	
		FROM usersUsage
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
