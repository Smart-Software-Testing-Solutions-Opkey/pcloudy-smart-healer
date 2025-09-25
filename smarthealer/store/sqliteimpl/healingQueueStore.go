package sqliteimpl

import (
	"context"
	"fmt"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store"
	"github.com/jmoiron/sqlx"
)

type sqliteHealingQueueStore struct {
	tx *sqlx.Tx
}

func NewSqliteHealingQueueStore(tx *sqlx.Tx) *sqliteHealingQueueStore {
	return &sqliteHealingQueueStore{
		tx: tx,
	}
}

func (s *sqliteHealingQueueStore) Add(ctx context.Context, infoJson, optJson string) error {
	query := `INSERT INTO healing_queue (info_json, opt_json) VALUES (?, ?);`

	stmt, err := s.tx.PreparexContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, infoJson, optJson)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}

func (s *sqliteHealingQueueStore) Remove(ctx context.Context, queueId int64) error {
	query := `DELETE FROM healing_queue WHERE id = ?;`

	stmt, err := s.tx.PreparexContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	r, err := stmt.ExecContext(ctx, queueId)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	rows, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows != 1 {
		return fmt.Errorf("expected to affect only 1 row, %d affected", rows)
	}

	return nil
}

func (s *sqliteHealingQueueStore) Length(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) from healing_queue;`

	var count int64
	if err := s.tx.GetContext(ctx, &count, query); err != nil {
		return -1, fmt.Errorf("failed to retrieve count: %w", err)
	}

	return count, nil
}

func (s *sqliteHealingQueueStore) GetOldestEntry(ctx context.Context) (*store.HealingQueueEntry, error) {
	query := `SELECT id, info_json, opt_json
	FROM healing_queue
	ORDER BY created_at ASC LIMIT 1;`

	p := struct {
		Id       int64  `db:"id"`
		InfoJson string `db:"info_json"`
		OptJson  string `db:"opt_json"`
	}{}
	if err := s.tx.GetContext(ctx, &p, query); err != nil {
		return nil, fmt.Errorf("failed to retrieve oldest healing entry: %w", err)
	}

	return &store.HealingQueueEntry{
		Id:       p.Id,
		InfoJson: p.InfoJson,
		OptJson:  p.OptJson,
	}, nil
}
