package sqlite

import (
	"context"
	"tgbot/common"
	"tgbot/internal/store"
)

func (d *DB) AiModelList(ctx context.Context) ([]*store.AiModel, error) {
	method := "AiModelList()"
	q := `SELECT * FROM aiModels`

	rows, err := d.db.QueryContext(ctx, q)
	if err != nil {
		return nil, common.WrapErrors(method, store.ErrDBQueryError, err)
	}
	defer closeRows(rows)

	list := make([]*store.AiModel, 0)
	for rows.Next() {
		var entity store.AiModel
		err := rows.Scan(
			&entity.ID,
			&entity.Title,
			&entity.APIName,
			&entity.ModelType)
		if err != nil {
			return nil, common.WrapErrors(method, store.ErrDBScanRowError, err)
		}
		list = append(list, &entity)
	}

	if err := rows.Err(); err != nil {
		return nil, common.WrapErrors(method, store.ErrDBRowError, err)
	}

	return list, nil
}
