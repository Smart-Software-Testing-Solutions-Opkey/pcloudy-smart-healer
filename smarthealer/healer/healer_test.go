package healer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/page"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/platform"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/retrieval"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store/sqliteimpl"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Mock intelligence system for testing
type mockIntelSystem struct{}

func (m *mockIntelSystem) GenerateElementDescription(ctx context.Context, pageSource, elementSource string) (string, error) {
	return "Mock description", nil
}

func (m *mockIntelSystem) GenerateLocator(ctx context.Context, desc string, root page.Page, plat platform.Platform) (string, error) {
	return "//button", nil
}

func (m *mockIntelSystem) CompareScreenShot(ctx context.Context, img1, img2 string) (bool, error) {
	return true, nil
}

func setupTestHealer(t *testing.T) (*Healer, *store.UnitOfWorkFactory, string, *sqlx.DB) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_healer.db")

	// Run migrations
	err := store.EnsureMigrations(dbPath)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Open database
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Create UnitOfWorkFactory
	uowF := store.NewUnitOfWorkFactory(db, store.FactoryParams{
		PageStoreFactory: func(tx *sqlx.Tx) store.PageStore {
			return sqliteimpl.NewSqlitePageStore(tx)
		},
		LocatorStoreFactory: func(tx *sqlx.Tx) store.LocatorStore {
			return sqliteimpl.NewSqliteLocatorStore(tx)
		},
		DescriptionQueueFactory: func(tx *sqlx.Tx) store.DescriptionQueue {
			return sqliteimpl.NewSqliteDescriptionQueueStore(tx)
		},
		HealingQueueFactory: func(tx *sqlx.Tx) store.HealingQueue {
			return sqliteimpl.NewSqliteHealingQueueStore(tx)
		},
	})

	// Create mock intelligence system
	mockIntel := &mockIntelSystem{}

	// Create page retriever
	pageRetriever := retrieval.NewPageRetriever(uowF, mockIntel)

	// Create background worker
	bg, err := NewBGWorker(mockIntel, uowF)
	if err != nil {
		t.Fatalf("Failed to create background worker: %v", err)
	}

	// Create healer
	healer := NewHealer(mockIntel, pageRetriever, uowF, bg)

	return healer, uowF, dbPath, db
}

func TestResolveLocator_NewEntry_CommitsToDatabase(t *testing.T) {
	healer, _, _, db := setupTestHealer(t)
	defer db.Close()

	ctx := context.Background()

	// Create test input
	info := LocatorInfo{
		ProjectId:  "test-project",
		PageSource: "<html><body><button id='test'>Click me</button></body></html>",
		B64Png:     "base64encodedimage",
		XPath:      "//button[@id='test']",
		ContextId:  "test-context",
		Platform:   platform.WebPlatform,
		PageType:   page.HTMLPageType,
	}

	opt := ResolveOptions{
		ComparisionMode: retrieval.AutomaticComparisionMode,
	}

	// Call ResolveLocator (should create new entry)
	result, err := healer.ResolveLocator(ctx, info, opt, nil)
	if err != nil {
		t.Fatalf("ResolveLocator failed: %v", err)
	}

	if result != info.XPath {
		t.Errorf("Expected result %s, got %s", info.XPath, result)
	}

	// Verify data was committed to database
	var pageCount int
	err = db.QueryRow("SELECT COUNT(*) FROM page WHERE project_id = ?", info.ProjectId).Scan(&pageCount)
	if err != nil {
		t.Fatalf("Failed to query page table: %v", err)
	}

	if pageCount != 1 {
		t.Errorf("Expected 1 page entry, got %d", pageCount)
	}

	// Verify locator was also saved
	var locatorCount int
	err = db.QueryRow("SELECT COUNT(*) FROM locator").Scan(&locatorCount)
	if err != nil {
		t.Fatalf("Failed to query locator table: %v", err)
	}

	if locatorCount != 1 {
		t.Errorf("Expected 1 locator entry, got %d", locatorCount)
	}

	t.Log("✓ ResolveLocator successfully committed new entry to database")
}

func TestResolveLocator_WithPassedTransaction_DoesNotCommit(t *testing.T) {
	healer, uowF, _, db := setupTestHealer(t)
	defer db.Close()

	ctx := context.Background()

	// Create a transaction
	uow, err := uowF.NewUnitOfWork(ctx)
	if err != nil {
		t.Fatalf("Failed to create unit of work: %v", err)
	}
	defer uow.Rollback()

	// Create test input
	info := LocatorInfo{
		ProjectId:  "test-project-2",
		PageSource: "<html><body><button id='test2'>Click me</button></body></html>",
		B64Png:     "base64encodedimage",
		XPath:      "//button[@id='test2']",
		ContextId:  "test-context-2",
		Platform:   platform.WebPlatform,
		PageType:   page.HTMLPageType,
	}

	opt := ResolveOptions{
		ComparisionMode: retrieval.AutomaticComparisionMode,
	}

	// Call ResolveLocator with passed transaction
	result, err := healer.ResolveLocator(ctx, info, opt, uow)
	if err != nil {
		t.Fatalf("ResolveLocator failed: %v", err)
	}

	if result != info.XPath {
		t.Errorf("Expected result %s, got %s", info.XPath, result)
	}

	// Data should NOT be visible yet (transaction not committed)
	var pageCount int
	err = db.QueryRow("SELECT COUNT(*) FROM page WHERE project_id = ?", info.ProjectId).Scan(&pageCount)
	if err != nil {
		t.Fatalf("Failed to query page table: %v", err)
	}

	if pageCount != 0 {
		t.Errorf("Expected 0 page entries (not committed), got %d", pageCount)
	}

	// Now commit the transaction
	err = uow.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Now data should be visible
	err = db.QueryRow("SELECT COUNT(*) FROM page WHERE project_id = ?", info.ProjectId).Scan(&pageCount)
	if err != nil {
		t.Fatalf("Failed to query page table after commit: %v", err)
	}

	if pageCount != 1 {
		t.Errorf("Expected 1 page entry after commit, got %d", pageCount)
	}

	t.Log("✓ ResolveLocator with passed transaction correctly delegates commit to caller")
}

func TestResolveLocator_InvalidXPath_RollsBack(t *testing.T) {
	healer, _, _, db := setupTestHealer(t)
	defer db.Close()

	ctx := context.Background()

	// Create test input with invalid XPath
	info := LocatorInfo{
		ProjectId:  "test-project-3",
		PageSource: "<html><body><button id='test'>Click me</button></body></html>",
		B64Png:     "base64encodedimage",
		XPath:      "//button[@id='nonexistent']", // This XPath doesn't exist in the page
		ContextId:  "test-context-3",
		Platform:   platform.WebPlatform,
		PageType:   page.HTMLPageType,
	}

	opt := ResolveOptions{
		ComparisionMode: retrieval.AutomaticComparisionMode,
	}

	// Call ResolveLocator (should fail validation)
	_, err := healer.ResolveLocator(ctx, info, opt, nil)
	if err == nil {
		t.Fatal("Expected error for invalid XPath, got nil")
	}

	// Verify no data was committed to database
	var pageCount int
	err = db.QueryRow("SELECT COUNT(*) FROM page WHERE project_id = ?", info.ProjectId).Scan(&pageCount)
	if err != nil {
		t.Fatalf("Failed to query page table: %v", err)
	}

	if pageCount != 0 {
		t.Errorf("Expected 0 page entries (should be rolled back), got %d", pageCount)
	}

	t.Log("✓ ResolveLocator correctly rolled back invalid entry")
}

func TestHandleNewEntry_DirectCall_VerifyDatabaseWrites(t *testing.T) {
	healer, uowF, _, db := setupTestHealer(t)
	defer db.Close()

	ctx := context.Background()

	// Create a transaction
	uow, err := uowF.NewUnitOfWork(ctx)
	if err != nil {
		t.Fatalf("Failed to create unit of work: %v", err)
	}
	defer uow.Rollback()

	// Create test input
	info := LocatorInfo{
		ProjectId:  "test-project-4",
		PageSource: "<html><body><div id='container'><button>Click</button></div></body></html>",
		B64Png:     "base64encodedimage",
		XPath:      "//div[@id='container']/button",
		ContextId:  "test-context-4",
		Platform:   platform.WebPlatform,
		PageType:   page.HTMLPageType,
	}

	// Call handleNewEntry directly
	result, err := healer.handleNewEntry(ctx, info, uow)
	if err != nil {
		t.Fatalf("handleNewEntry failed: %v", err)
	}

	if result != info.XPath {
		t.Errorf("Expected result %s, got %s", info.XPath, result)
	}

	// Commit the transaction
	err = uow.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Verify page was inserted
	var pageID int
	var pageSource, locator, platform, pageType, contextID, projectID string
	err = db.QueryRow(`
		SELECT page_id, page_source, locator, platform, page_type, context_id, project_id
		FROM page WHERE project_id = ?
	`, info.ProjectId).Scan(&pageID, &pageSource, &locator, &platform, &pageType, &contextID, &projectID)
	if err != nil {
		t.Fatalf("Failed to query page: %v", err)
	}

	if pageID == 0 {
		t.Error("Expected non-zero page_id")
	}
	if pageSource != info.PageSource {
		t.Errorf("Expected page_source %s, got %s", info.PageSource, pageSource)
	}
	if projectID != info.ProjectId {
		t.Errorf("Expected project_id %s, got %s", info.ProjectId, projectID)
	}

	// Verify locator was inserted
	var locatorID int
	var locatorXPath string
	err = db.QueryRow(`
		SELECT locator_id, locator FROM locator WHERE page_id = ?
	`, pageID).Scan(&locatorID, &locatorXPath)
	if err != nil {
		t.Fatalf("Failed to query locator: %v", err)
	}

	if locatorID == 0 {
		t.Error("Expected non-zero locator_id")
	}
	if locatorXPath != info.XPath {
		t.Errorf("Expected locator %s, got %s", info.XPath, locatorXPath)
	}

	// Verify description queue entry was added
	var descQueueCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM description_queue WHERE page_id = ? AND locator_id = ?
	`, pageID, locatorID).Scan(&descQueueCount)
	if err != nil {
		t.Fatalf("Failed to query description_queue: %v", err)
	}

	if descQueueCount != 1 {
		t.Errorf("Expected 1 description_queue entry, got %d", descQueueCount)
	}

	t.Log("✓ handleNewEntry correctly writes all data to database")
}

func TestBackgroundWorker_HealWorkerFunc_CommitsAfterSuccess(t *testing.T) {
	_, uowF, _, db := setupTestHealer(t)
	defer db.Close()

	ctx := context.Background()

	// Create background worker
	mockIntel := &mockIntelSystem{}
	bg, err := NewBGWorker(mockIntel, uowF)
	if err != nil {
		t.Fatalf("Failed to create background worker: %v", err)
	}

	// Add a healing queue entry
	uow, err := uowF.NewUnitOfWork(ctx)
	if err != nil {
		t.Fatalf("Failed to create unit of work: %v", err)
	}

	err = uow.HealingQueue.Add(ctx, `{"ProjectId":"test","PageSource":"<html><body><button id='test'>Test</button></body></html>","B64Png":"img","XPath":"//button[@id='test']","ContextId":"ctx","Platform":"Web","PageType":"HTMLPageType"}`, `{"ComparisionMode":"Automatic"}`)
	if err != nil {
		t.Fatalf("Failed to add healing queue entry: %v", err)
	}

	err = uow.Commit()
	if err != nil {
		t.Fatalf("Failed to commit healing queue entry: %v", err)
	}

	// Create a mock resolver that succeeds
	mockResolver := func(ctx context.Context, info LocatorInfo, opt ResolveOptions, u *store.UnitOfWork) (string, error) {
		// Simulate adding a page entry
		pageID, err := u.Pages.Add(ctx, store.PageEntry{
			PageSource: info.PageSource,
			Locator:    info.XPath,
			B64Png:     info.B64Png,
			ContextId:  info.ContextId,
			ProjectId:  info.ProjectId,
			Platform:   info.Platform,
			PageType:   info.PageType,
		})
		if err != nil {
			return "", err
		}

		_, err = u.Locators.Add(ctx, store.LocatorEntry{
			PageId:      pageID,
			Locator:     info.XPath,
			Description: "test",
		})
		if err != nil {
			return "", err
		}

		return info.XPath, nil
	}

	// Create heal worker function
	healWorker := bg.HealWorkerFunc(mockResolver)

	// Execute the worker
	err = healWorker(ctx)
	if err != nil {
		t.Fatalf("HealWorkerFunc failed: %v", err)
	}

	// Verify the page was committed to database
	var pageCount int
	err = db.QueryRow("SELECT COUNT(*) FROM page WHERE project_id = ?", "test").Scan(&pageCount)
	if err != nil {
		t.Fatalf("Failed to query page table: %v", err)
	}

	if pageCount != 1 {
		t.Errorf("Expected 1 page entry after heal worker, got %d", pageCount)
	}

	// Verify healing queue entry was removed
	var queueCount int
	err = db.QueryRow("SELECT COUNT(*) FROM healing_queue").Scan(&queueCount)
	if err != nil {
		t.Fatalf("Failed to query healing_queue: %v", err)
	}

	if queueCount != 0 {
		t.Errorf("Expected 0 healing_queue entries after processing, got %d", queueCount)
	}

	t.Log("✓ HealWorkerFunc correctly commits data and removes queue entry")
}
