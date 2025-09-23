package sqliteimpl

import (
	"context"
	"fmt"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/page"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store"
	"github.com/jmoiron/sqlx"
)

type sqlitePageStore struct {
	db sqlx.DB
}

func NewSqlitePageStore(db sqlx.DB) *sqlitePageStore {
	return &sqlitePageStore{
		db: db,
	}
}

func (s *sqlitePageStore) Add(ctx context.Context, entry store.PageEntry) (int, error) {
	query := `INSERT INTO 
	page(page_source, locator, b64_png, context_id, project_id, platform, page_type)
	VALUES (?, ?, ?, ?, ?, ?, ?);`

	stmt, err := s.db.PreparexContext(ctx, query)
	if err != nil {
		return -1, fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	r, err := stmt.ExecContext(ctx,
		entry.PageSource,
		entry.Locator,
		entry.B64Png,
		entry.ContextId,
		entry.ProjectId,
		entry.Platform.String(),
		entry.PageType.String(),
	)
	if err != nil {
		return -1, fmt.Errorf("failed to insert page: %w", err)
	}

	id, err := r.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("failed to get the latest page id: %w", err)
	}

	return int(id), nil
}

func (s *sqlitePageStore) GetPageSourceInfo(ctx context.Context, pageId int) (store.PageSrcInfo, error) {
	query := `SELECT page_source, page_type FROM page WHERE page_id = ?;`

	type pageInfo struct {
		PageSource string `db:"page_source"`
		PageType   string `db:"page_type"`
	}
	var p pageInfo
	if err := s.db.Get(&p, query, pageId); err != nil {
		return store.PageSrcInfo{}, fmt.Errorf("failed to query page for pageId(%d): %w", pageId, err)
	}

	return store.PageSrcInfo{
		PageSource: p.PageSource,
		PageType:   page.NewPageTypeFromString(p.PageType),
	}, nil
}

func (s *sqlitePageStore) GetPagePNG(ctx context.Context, pageId int) (string, error) {
	query := `SELECT b64_png FROM page WHERE page_id = ?;`

	var str string
	if err := s.db.Get(&str, query, pageId); err != nil {
		return "", fmt.Errorf("failed to query png for pageId(%d): %w", pageId, err)
	}

	return str, nil
}

func (s *sqlitePageStore) GetFirstPageWithContext(ctx context.Context, projectId, locator, contextId string) (int, error) {
	query := `SELECT page_id FROM page WHERE project_id = ? AND locator = ? AND context_id = ?;`

	var id int
	if err := s.db.Get(&id, query, projectId, locator, contextId); err != nil {
		return -1, fmt.Errorf("failed to query pageId for given params: %w", err)
	}

	return id, nil
}

func (s *sqlitePageStore) GetPages(ctx context.Context, projectId, locator string) ([]int, error) {
	query := `SELECT page_id FROM page where project_id = ? AND locator = ?;`

	var ids []int
	if err := s.db.Select(&ids, query, projectId, locator); err != nil {
		return nil, fmt.Errorf("failed to query pageIds for given params: %w", err)
	}

	return ids, nil
}
