package sqlite

import (
	"context"
	"fmt"
	"strings"
	"tgbot/common"
	"tgbot/internal/store"
	"time"
)

func (d *DB) UserCreate(ctx context.Context, entity *store.User) (*store.User, error) {
	fields := []string{"id", "chatModelId", "imageModelId", "tariffId", "lastLimitReset"}
	args := []any{entity.ID, entity.ChatModelID, entity.ImageModelID, entity.TariffID, common.TimeNowUTCDay()}

	q := "INSERT INTO users (" + strings.Join(fields, ", ") + ") VALUES (" + placeholders(len(fields)) + ");\n"
	q += "INSERT INTO usersUsage (userId, aiModelId, count) VALUES (" + placeholdersRange(len(fields)+1, 3) + ")"

	args = append(args, entity.ID, entity.ChatModelID, 0, common.TimeNowUTCDay())

	if _, err := d.db.ExecContext(ctx, q, args...); err != nil {
		return nil, common.WrapErrors("UserCreate()", store.ErrDBQueryError, err)
	}

	return entity, nil
}

func (d *DB) UserList(ctx context.Context, filter *store.UserFilter) ([]*store.User, error) {
	method := "UserList()"
	where, args := []string{"1 = 1"}, []any{}

	if filter.ID != nil {
		where, args = append(where, "id = ?"), append(args, filter.ID)
	}
	if filter.ChatModelID != nil {
		where, args = append(where, "chatModelId = ?"), append(args, filter.ChatModelID)
	}
	if filter.ImageModelID != nil {
		where, args = append(where, "imageModelId = ?"), append(args, filter.ImageModelID)
	}
	if filter.TariffID != nil {
		where, args = append(where, "tariffId = ?"), append(args, filter.TariffID)
	}
	if filter.LastLimitReset != nil {
		where, args = append(where, "lastLimitReset = ?"), append(args, filter.LastLimitReset)
	}
	if filter.SelfBlock != nil {
		where, args = append(where, "selfBlock = ?"), append(args, filter.SelfBlock)
	}
	if filter.Blocked != nil {
		where, args = append(where, "blocked = ?"), append(args, filter.Blocked)
	}
	if filter.BlockReason != nil {
		where, args = append(where, "blockReason = ?"), append(args, filter.BlockReason)
	}
	if filter.SkipNewDialogMessage != nil {
		where, args = append(where, "skipNewDialogMessage = ?"), append(args, filter.SkipNewDialogMessage)
	}

	q := `
		SELECT *		
		FROM users
		WHERE ` + strings.Join(where, " AND ")

	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, common.WrapErrors(method, store.ErrDBQueryError, err)
	}
	defer closeRows(rows)

	list := make([]*store.User, 0)
	for rows.Next() {
		var entity store.User
		var lastLimitReset string
		if err := rows.Scan(
			&entity.ID,
			&entity.ChatModelID,
			&entity.ImageModelID,
			&entity.TariffID,
			&lastLimitReset,
			&entity.SelfBlock,
			&entity.Blocked,
			&entity.BlockReason,
			&entity.SkipNewDialogMessage,
		); err != nil {
			return nil, common.WrapErrors(method, store.ErrDBScanRowError, err)
		}
		entity.LastLimitReset, err = time.Parse(dateLayout(), lastLimitReset)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to parse lastLimitReset: %w", method, err)
		}
		list = append(list, &entity)
	}

	if err := rows.Err(); err != nil {
		return nil, common.WrapErrors(method, store.ErrDBRowError, err)
	}

	return list, nil
}

func (d *DB) UserUpdate(ctx context.Context, entity *store.User) (*store.User, error) {
	q := `UPDATE users
			SET
				chatModelId = ?,
				imageModelId = ?,
				tariffId = ?,
				lastLimitReset = ?,
				selfBlock = ?,
				blocked = ?,
				blockReason = ?,
				skipNewDialogMessage = ?
			WHERE
				id = ?;`
	_, err := d.db.ExecContext(ctx, q,
		entity.ChatModelID,
		entity.ImageModelID,
		entity.TariffID,
		common.TimeNowUTCDay(),
		entity.SelfBlock,
		entity.Blocked,
		entity.BlockReason,
		entity.SkipNewDialogMessage,
		entity.ID)
	if err != nil {
		return nil, common.WrapErrors("UserUpdate()", store.ErrDBQueryError, err)
	}
	return entity, nil
}

func (d *DB) UserDelete(ctx context.Context, filter *store.UserFilter) error {
	method := "UserDelete()"
	where, args := []string{}, []any{}

	if filter.ID != nil {
		where, args = append(where, "id = ?"), append(args, filter.ID)
	}
	if filter.ChatModelID != nil {
		where, args = append(where, "chatModelId = ?"), append(args, filter.ChatModelID)
	}
	if filter.ImageModelID != nil {
		where, args = append(where, "imageModelId = ?"), append(args, filter.ImageModelID)
	}
	if filter.TariffID != nil {
		where, args = append(where, "tariffId = ?"), append(args, filter.TariffID)
	}
	if filter.LastLimitReset != nil {
		where, args = append(where, "lastLimitReset = ?"), append(args, filter.LastLimitReset)
	}
	if filter.SelfBlock != nil {
		where, args = append(where, "selfBlock = ?"), append(args, filter.SelfBlock)
	}
	if filter.Blocked != nil {
		where, args = append(where, "blocked = ?"), append(args, filter.Blocked)
	}
	if filter.BlockReason != nil {
		where, args = append(where, "blockReason = ?"), append(args, filter.BlockReason)
	}
	if filter.SkipNewDialogMessage != nil {
		where, args = append(where, "skipNewDialogMessage = ?"), append(args, filter.SkipNewDialogMessage)
	}

	if len(where) == 0 {
		return common.WrapErrors(method, store.ErrDBNoFilterProvided)
	}

	q := `
		DELETE	
		FROM users
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
