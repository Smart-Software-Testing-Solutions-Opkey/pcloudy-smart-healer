package sqliteimpl

import (
	"context"
	"fmt"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store"
	"github.com/jmoiron/sqlx"
)

type sqliteDescriptionQueueStore struct {
	tx *sqlx.Tx
}

func NewSqliteDescriptionQueueStore(tx *sqlx.Tx) *sqliteDescriptionQueueStore {
	return &sqliteDescriptionQueueStore{
		tx: tx,
	}
}

func (s *sqliteDescriptionQueueStore) Add(ctx context.Context, locatorId, pageId int) error {
	query := `INSERT INTO description_queue (page_id, locator_id) VALUES (?, ?);`

	stmt, err := s.tx.PreparexContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, pageId, locatorId)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}

func (s *sqliteDescriptionQueueStore) Remove(ctx context.Context, locatorId, pageId int) error {
	query := `DELETE FROM description_queue WHRE page_id = ? AND locator_id = ?;`

	stmt, err := s.tx.PreparexContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	r, err := stmt.ExecContext(ctx, pageId, locatorId)
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

func (s *sqliteDescriptionQueueStore) Length(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM description_queue;`

	var count int64
	if err := s.tx.GetContext(ctx, &count, query); err != nil {
		return -1, fmt.Errorf("failed to retreive count: %w", err)
	}

	return count, nil
}

func (s *sqliteDescriptionQueueStore) GetOldestEntry(ctx context.Context) (*store.DescriptionQueueEntry, error) {
	query := `SELECT page_id, locator_id 
	FROM description_queue
	ORDER BY created_at ASC LIMIT 1;`

	p := struct {
		PageId    int `db:"page_id"`
		LocatorId int `db:"locator_id"`
	}{}
	if err := s.tx.GetContext(ctx, &p, query); err != nil {
		return nil, fmt.Errorf("failed to retreive oldest locator: %w", err)
	}

	return &store.DescriptionQueueEntry{
		PageId:    p.PageId,
		LocatorId: p.LocatorId,
	}, nil
}

var _ store.DescriptionQueue = &sqliteDescriptionQueueStore{}
