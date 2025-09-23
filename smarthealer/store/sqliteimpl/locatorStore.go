package sqliteimpl

import (
	"context"
	"fmt"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store"
	"github.com/jmoiron/sqlx"
)

type sqliteLocatorStore struct {
	db sqlx.DB
}

func NewSqliteLocatorStore(db sqlx.DB) *sqlitePageStore {
	return &sqlitePageStore{
		db: db,
	}
}

func (s *sqliteLocatorStore) Add(ctx context.Context, entry store.LocatorEntry) (int, error) {
	query := `INSERT INTO locator (page_id, locator, description) 
	VALUES (?, ?, ?);`

	stmt, err := s.db.PreparexContext(ctx, query)
	if err != nil {
		return -1, fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	r, err := stmt.ExecContext(ctx, entry.PageId, entry.Locator, entry.Description)
	if err != nil {
		return -1, fmt.Errorf("failed to execute statement: %w", err)
	}

	id, err := r.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("failed to retreive last insert id: %w", err)
	}

	return int(id), nil
}

func (s *sqliteLocatorStore) UpdateDescription(ctx context.Context, locatorId int, desc string) error {
	query := `UPDATE locator SET description = ? WHERE locator_id = ?;`

	stmt, err := s.db.PreparexContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, desc, locatorId)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}

func (s *sqliteLocatorStore) GetPageLocators(ctx context.Context, pageId int) ([]string, error) {
	query := `SELECT locator FROM locator WHERE page_id = ? ORDER BY created_at DESC;`

	var ids []string
	if err := s.db.Select(&ids, query, pageId); err != nil {
		return nil, fmt.Errorf("failed to query locators: %w", err)
	}

	return ids, nil
}

func (s *sqliteLocatorStore) GetLatestPageDescription(ctx context.Context, pageId int) (string, error) {
	query := `SELECT description FROM locator 
	WHERE page_id = ? ORDER BY created_at DESC
	LIMIT 1;`

	var desc string
	if err := s.db.Get(&desc, query, pageId); err != nil {
		return "", fmt.Errorf("failed to query latest description: %w", err)
	}

	return desc, nil
}

func (s *sqliteLocatorStore) GetLocator(ctx context.Context, locatorId int) (string, error) {
	query := `SELECT locator from locator WHERE locator_id = ?;`

	var locator string
	if err := s.db.Get(&locator, query, locatorId); err != nil {
		return "", fmt.Errorf("failed to query locator: %w", err)
	}

	return locator, nil
}
