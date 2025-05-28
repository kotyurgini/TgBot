package sqlite

import (
	"context"
	"fmt"
	"strings"
	"tgbot/common"
	"tgbot/internal/store"
	"time"
)

func (d *DB) ChatMessageCreate(ctx context.Context, entity *store.ChatMessage) (*store.ChatMessage, error) {
	fields := []string{"dialogId", "\"order\"", "\"role\"", "content", "created"}
	args := []any{entity.DialogID, entity.Order, entity.Role, entity.Content, entity.Created}

	q := "INSERT INTO chatMessages (" + strings.Join(fields, ", ") + ") VALUES (" + placeholders(len(fields)) + ") RETURNING id"

	if err := d.db.QueryRowContext(ctx, q, args...).Scan(
		&entity.ID,
	); err != nil {
		return nil, common.WrapErrors("ChatMessageCreate()", store.ErrDBQueryError, err)
	}

	return entity, nil
}

func (d *DB) ChatMessageList(ctx context.Context, filter *store.ChatMessageFilter) ([]*store.ChatMessage, error) {
	method := "ChatMessageList()"
	where, args := []string{"1 = 1"}, []any{}

	if filter.ID != nil {
		where, args = append(where, "id = ?"), append(args, filter.ID)
	}
	if filter.DialogID != nil {
		where, args = append(where, "dialogId = ?"), append(args, filter.DialogID)
	}
	if filter.Order != nil {
		where, args = append(where, "\"order\" = ?"), append(args, filter.Order)
	}
	if filter.Role != nil {
		where, args = append(where, "\"role\" = ?"), append(args, filter.Role)
	}
	if filter.Content != nil {
		where, args = append(where, "content = ?"), append(args, filter.Content)
	}
	if filter.Created != nil {
		where, args = append(where, "created = ?"), append(args, filter.Created)
	}

	q := `
		SELECT *	
		FROM chatMessages
		WHERE ` + strings.Join(where, " AND ")

	rows, err := d.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, common.WrapErrors(method, store.ErrDBQueryError, err)
	}
	defer closeRows(rows)

	list := make([]*store.ChatMessage, 0)
	for rows.Next() {
		var entity store.ChatMessage
		var created string
		if err := rows.Scan(
			&entity.ID, &entity.DialogID, &entity.Order, &entity.Role, &entity.Content, &created,
		); err != nil {
			return nil, common.WrapErrors(method, store.ErrDBScanRowError, err)
		}

		entity.Created, err = time.Parse(dateLayout(), created)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to parse created: %w", method, err)
		}
		list = append(list, &entity)
	}

	if err := rows.Err(); err != nil {
		return nil, common.WrapErrors(method, store.ErrDBRowError, err)
	}

	return list, nil
}

func (d *DB) ChatMessageUpdate(ctx context.Context, entity *store.ChatMessage) (*store.ChatMessage, error) {
	q := `UPDATE chatMessages
			SET
				dialogId = ?,
				"order" = ?,
				"role" = ?,
				content = ?,
				created = ?
			WHERE
				id = ?;`

	_, err := d.db.ExecContext(ctx, q,
		entity.DialogID,
		entity.Order,
		entity.Role,
		entity.Content,
		entity.Created,
		entity.ID)
	if err != nil {
		return nil, common.WrapErrors("ChatMessageUpdate()", store.ErrDBQueryError, err)
	}
	return entity, nil
}

func (d *DB) ChatMessageDelete(ctx context.Context, filter *store.ChatMessageFilter) error {
	method := "ChatMessageDelete()"
	where, args := []string{}, []any{}

	if filter.ID != nil {
		where, args = append(where, "id = ?"), append(args, filter.ID)
	}
	if filter.DialogID != nil {
		where, args = append(where, "dialogId = ?"), append(args, filter.DialogID)
	}
	if filter.Order != nil {
		where, args = append(where, "\"order\" = ?"), append(args, filter.Order)
	}
	if filter.Role != nil {
		where, args = append(where, "\"role\" = ?"), append(args, filter.Role)
	}
	if filter.Content != nil {
		where, args = append(where, "content = ?"), append(args, filter.Content)
	}
	if filter.Created != nil {
		where, args = append(where, "created = ?"), append(args, filter.Created)
	}

	if len(where) == 0 {
		return common.WrapErrors(method, store.ErrDBNoFilterProvided)
	}

	q := `
		DELETE	
		FROM chatMessages
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
