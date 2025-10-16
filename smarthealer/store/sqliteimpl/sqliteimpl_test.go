package sqliteimpl

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/page"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/platform"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sqlx.DB, string) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

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

	return db, dbPath
}

// ============================================================================
// PageStore Tests
// ============================================================================

func TestPageStore_Add(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	pageStore := NewSqlitePageStore(tx)

	entry := store.PageEntry{
		PageSource: "<html><body><button id='test'>Click</button></body></html>",
		Locator:    "//button[@id='test']",
		B64Png:     "base64imagedata",
		ContextId:  "test-context",
		ProjectId:  "test-project",
		Platform:   platform.WebPlatform,
		PageType:   page.HTMLPageType,
	}

	pageId, err := pageStore.Add(ctx, entry)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if pageId <= 0 {
		t.Errorf("Expected positive page ID, got %d", pageId)
	}

	// Verify the data was inserted
	var count int
	err = tx.Get(&count, "SELECT COUNT(*) FROM page WHERE page_id = ?", pageId)
	if err != nil {
		t.Fatalf("Failed to verify insert: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 page entry, got %d", count)
	}

	// Commit and verify persistence
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	var persistedCount int
	err = db.Get(&persistedCount, "SELECT COUNT(*) FROM page WHERE page_id = ?", pageId)
	if err != nil {
		t.Fatalf("Failed to verify persisted data: %v", err)
	}

	if persistedCount != 1 {
		t.Errorf("Expected 1 persisted page entry, got %d", persistedCount)
	}

	t.Log("✓ PageStore.Add successfully inserts and persists data")
}

func TestPageStore_GetPageSourceInfo(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	pageStore := NewSqlitePageStore(tx)

	// Insert a page
	entry := store.PageEntry{
		PageSource: "<html><body><div>Test</div></body></html>",
		Locator:    "//div",
		B64Png:     "imagedata",
		ContextId:  "ctx",
		ProjectId:  "proj",
		Platform:   platform.AndroidPlatform,
		PageType:   page.XMLPageType,
	}

	pageId, err := pageStore.Add(ctx, entry)
	if err != nil {
		t.Fatalf("Failed to add page: %v", err)
	}

	// Retrieve page source info
	info, err := pageStore.GetPageSourceInfo(ctx, pageId)
	if err != nil {
		t.Fatalf("GetPageSourceInfo failed: %v", err)
	}

	if info.PageSource != entry.PageSource {
		t.Errorf("Expected PageSource %s, got %s", entry.PageSource, info.PageSource)
	}

	if info.PageType != entry.PageType {
		t.Errorf("Expected PageType %v, got %v", entry.PageType, info.PageType)
	}

	t.Log("✓ PageStore.GetPageSourceInfo retrieves correct data")
}

func TestPageStore_GetFirstPageWithContext(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	pageStore := NewSqlitePageStore(tx)

	// Insert multiple pages
	entries := []store.PageEntry{
		{
			PageSource: "<html>Page 1</html>",
			Locator:    "//button",
			B64Png:     "img1",
			ContextId:  "ctx-1",
			ProjectId:  "proj-1",
			Platform:   platform.WebPlatform,
			PageType:   page.HTMLPageType,
		},
		{
			PageSource: "<html>Page 2</html>",
			Locator:    "//button",
			B64Png:     "img2",
			ContextId:  "ctx-1",
			ProjectId:  "proj-1",
			Platform:   platform.WebPlatform,
			PageType:   page.HTMLPageType,
		},
	}

	firstPageId := -1
	for i, entry := range entries {
		pageId, err := pageStore.Add(ctx, entry)
		if err != nil {
			t.Fatalf("Failed to add page %d: %v", i, err)
		}
		if i == 0 {
			firstPageId = pageId
		}
	}

	// Query for first page
	retrievedId, err := pageStore.GetFirstPageWithContext(ctx, "proj-1", "//button", "ctx-1")
	if err != nil {
		t.Fatalf("GetFirstPageWithContext failed: %v", err)
	}

	if retrievedId != firstPageId {
		t.Errorf("Expected page ID %d, got %d", firstPageId, retrievedId)
	}

	t.Log("✓ PageStore.GetFirstPageWithContext returns correct page")
}

func TestPageStore_GetPages(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	pageStore := NewSqlitePageStore(tx)

	// Insert pages with same project and locator
	for i := 0; i < 3; i++ {
		entry := store.PageEntry{
			PageSource: "<html>Page</html>",
			Locator:    "//div",
			B64Png:     "img",
			ContextId:  "ctx",
			ProjectId:  "proj-test",
			Platform:   platform.WebPlatform,
			PageType:   page.HTMLPageType,
		}
		_, err := pageStore.Add(ctx, entry)
		if err != nil {
			t.Fatalf("Failed to add page: %v", err)
		}
	}

	// Get all pages
	pages, err := pageStore.GetPages(ctx, "proj-test", "//div")
	if err != nil {
		t.Fatalf("GetPages failed: %v", err)
	}

	if len(pages) != 3 {
		t.Errorf("Expected 3 pages, got %d", len(pages))
	}

	t.Log("✓ PageStore.GetPages returns all matching pages")
}

// ============================================================================
// LocatorStore Tests
// ============================================================================

func TestLocatorStore_Add(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// First create a page
	pageStore := NewSqlitePageStore(tx)
	pageEntry := store.PageEntry{
		PageSource: "<html>Test</html>",
		Locator:    "//button",
		B64Png:     "img",
		ContextId:  "ctx",
		ProjectId:  "proj",
		Platform:   platform.WebPlatform,
		PageType:   page.HTMLPageType,
	}
	pageId, err := pageStore.Add(ctx, pageEntry)
	if err != nil {
		t.Fatalf("Failed to add page: %v", err)
	}

	// Add locator
	locatorStore := NewSqliteLocatorStore(tx)
	locatorEntry := store.LocatorEntry{
		PageId:      pageId,
		Locator:     "//button[@id='submit']",
		Description: "Submit button",
	}

	locatorId, err := locatorStore.Add(ctx, locatorEntry)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if locatorId <= 0 {
		t.Errorf("Expected positive locator ID, got %d", locatorId)
	}

	// Verify insertion
	var count int
	err = tx.Get(&count, "SELECT COUNT(*) FROM locator WHERE locator_id = ?", locatorId)
	if err != nil {
		t.Fatalf("Failed to verify insert: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 locator entry, got %d", count)
	}

	// Commit and verify persistence
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	var persistedCount int
	err = db.Get(&persistedCount, "SELECT COUNT(*) FROM locator WHERE locator_id = ?", locatorId)
	if err != nil {
		t.Fatalf("Failed to verify persisted data: %v", err)
	}

	if persistedCount != 1 {
		t.Errorf("Expected 1 persisted locator entry, got %d", persistedCount)
	}

	t.Log("✓ LocatorStore.Add successfully inserts and persists data")
}

func TestLocatorStore_UpdateDescription(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Create page and locator
	pageStore := NewSqlitePageStore(tx)
	pageId, _ := pageStore.Add(ctx, store.PageEntry{
		PageSource: "<html>Test</html>",
		Locator:    "//button",
		B64Png:     "img",
		ContextId:  "ctx",
		ProjectId:  "proj",
		Platform:   platform.WebPlatform,
		PageType:   page.HTMLPageType,
	})

	locatorStore := NewSqliteLocatorStore(tx)
	locatorId, _ := locatorStore.Add(ctx, store.LocatorEntry{
		PageId:      pageId,
		Locator:     "//button",
		Description: "Old description",
	})

	// Update description
	newDesc := "New updated description"
	err = locatorStore.UpdateDescription(ctx, locatorId, newDesc)
	if err != nil {
		t.Fatalf("UpdateDescription failed: %v", err)
	}

	// Verify update
	var desc string
	err = tx.Get(&desc, "SELECT description FROM locator WHERE locator_id = ?", locatorId)
	if err != nil {
		t.Fatalf("Failed to query description: %v", err)
	}

	if desc != newDesc {
		t.Errorf("Expected description %s, got %s", newDesc, desc)
	}

	// Commit and verify persistence
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	var persistedDesc string
	err = db.Get(&persistedDesc, "SELECT description FROM locator WHERE locator_id = ?", locatorId)
	if err != nil {
		t.Fatalf("Failed to verify persisted description: %v", err)
	}

	if persistedDesc != newDesc {
		t.Errorf("Expected persisted description %s, got %s", newDesc, persistedDesc)
	}

	t.Log("✓ LocatorStore.UpdateDescription successfully updates and persists")
}

func TestLocatorStore_GetPageLocators(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Create page
	pageStore := NewSqlitePageStore(tx)
	pageId, _ := pageStore.Add(ctx, store.PageEntry{
		PageSource: "<html>Test</html>",
		Locator:    "//button",
		B64Png:     "img",
		ContextId:  "ctx",
		ProjectId:  "proj",
		Platform:   platform.WebPlatform,
		PageType:   page.HTMLPageType,
	})

	// Add multiple locators
	locatorStore := NewSqliteLocatorStore(tx)
	expectedLocators := []string{
		"//button[@id='first']",
		"//button[@id='second']",
		"//button[@id='third']",
	}

	for _, loc := range expectedLocators {
		_, err := locatorStore.Add(ctx, store.LocatorEntry{
			PageId:      pageId,
			Locator:     loc,
			Description: "Test",
		})
		if err != nil {
			t.Fatalf("Failed to add locator: %v", err)
		}
	}

	// Get locators
	locators, err := locatorStore.GetPageLocators(ctx, pageId)
	if err != nil {
		t.Fatalf("GetPageLocators failed: %v", err)
	}

	if len(locators) != len(expectedLocators) {
		t.Errorf("Expected %d locators, got %d", len(expectedLocators), len(locators))
	}

	// Verify all expected locators are present (order may vary based on timestamp precision)
	locatorMap := make(map[string]bool)
	for _, loc := range locators {
		locatorMap[loc] = true
	}

	for _, expected := range expectedLocators {
		if !locatorMap[expected] {
			t.Errorf("Expected locator %s not found in results", expected)
		}
	}

	t.Log("✓ LocatorStore.GetPageLocators returns all locators")
}

// ============================================================================
// DescriptionQueue Tests
// ============================================================================

func TestDescriptionQueue_AddAndGetOldest(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	descQueue := NewSqliteDescriptionQueueStore(tx)

	// Add entries
	err = descQueue.Add(ctx, 1, 1)
	if err != nil {
		t.Fatalf("Failed to add to description queue: %v", err)
	}

	err = descQueue.Add(ctx, 2, 2)
	if err != nil {
		t.Fatalf("Failed to add second entry: %v", err)
	}

	// Get oldest entry
	entry, err := descQueue.GetOldestEntry(ctx)
	if err != nil {
		t.Fatalf("GetOldestEntry failed: %v", err)
	}

	if entry.LocatorId != 1 || entry.PageId != 1 {
		t.Errorf("Expected entry (1,1), got (%d,%d)", entry.LocatorId, entry.PageId)
	}

	// Commit and verify persistence
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM description_queue")
	if err != nil {
		t.Fatalf("Failed to verify persistence: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 persisted entries, got %d", count)
	}

	t.Log("✓ DescriptionQueue.Add and GetOldestEntry work correctly")
}

func TestDescriptionQueue_Remove(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	descQueue := NewSqliteDescriptionQueueStore(tx)

	// Add entry
	err = descQueue.Add(ctx, 10, 20)
	if err != nil {
		t.Fatalf("Failed to add: %v", err)
	}

	// Remove entry
	err = descQueue.Remove(ctx, 10, 20)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify removal
	var count int
	err = tx.Get(&count, "SELECT COUNT(*) FROM description_queue WHERE locator_id = 10 AND page_id = 20")
	if err != nil {
		t.Fatalf("Failed to verify removal: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 entries after removal, got %d", count)
	}

	// Commit and verify persistence
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	var persistedCount int
	err = db.Get(&persistedCount, "SELECT COUNT(*) FROM description_queue WHERE locator_id = 10 AND page_id = 20")
	if err != nil {
		t.Fatalf("Failed to verify persisted removal: %v", err)
	}

	if persistedCount != 0 {
		t.Errorf("Expected 0 persisted entries, got %d", persistedCount)
	}

	t.Log("✓ DescriptionQueue.Remove successfully removes and persists deletion")
}

// ============================================================================
// HealingQueue Tests
// ============================================================================

func TestHealingQueue_AddAndGetOldest(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	healingQueue := NewSqliteHealingQueueStore(tx)

	infoJson := `{"xpath":"//button"}`
	optJson := `{"mode":"auto"}`

	// Add entry
	err = healingQueue.Add(ctx, infoJson, optJson)
	if err != nil {
		t.Fatalf("Failed to add to healing queue: %v", err)
	}

	// Get oldest entry
	entry, err := healingQueue.GetOldestEntry(ctx)
	if err != nil {
		t.Fatalf("GetOldestEntry failed: %v", err)
	}

	if entry.InfoJson != infoJson {
		t.Errorf("Expected InfoJson %s, got %s", infoJson, entry.InfoJson)
	}

	if entry.OptJson != optJson {
		t.Errorf("Expected OptJson %s, got %s", optJson, entry.OptJson)
	}

	if entry.Id <= 0 {
		t.Errorf("Expected positive ID, got %d", entry.Id)
	}

	// Commit and verify persistence
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM healing_queue")
	if err != nil {
		t.Fatalf("Failed to verify persistence: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 persisted entry, got %d", count)
	}

	t.Log("✓ HealingQueue.Add and GetOldestEntry work correctly")
}

func TestHealingQueue_Remove(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	healingQueue := NewSqliteHealingQueueStore(tx)

	// Add entry
	err = healingQueue.Add(ctx, `{"test":"data"}`, `{"opt":"val"}`)
	if err != nil {
		t.Fatalf("Failed to add: %v", err)
	}

	// Get the entry to get its ID
	entry, err := healingQueue.GetOldestEntry(ctx)
	if err != nil {
		t.Fatalf("GetOldestEntry failed: %v", err)
	}

	// Remove entry
	err = healingQueue.Remove(ctx, entry.Id)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify removal
	var count int
	err = tx.Get(&count, "SELECT COUNT(*) FROM healing_queue WHERE id = ?", entry.Id)
	if err != nil {
		t.Fatalf("Failed to verify removal: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 entries after removal, got %d", count)
	}

	// Commit and verify persistence
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	var persistedCount int
	err = db.Get(&persistedCount, "SELECT COUNT(*) FROM healing_queue WHERE id = ?", entry.Id)
	if err != nil {
		t.Fatalf("Failed to verify persisted removal: %v", err)
	}

	if persistedCount != 0 {
		t.Errorf("Expected 0 persisted entries, got %d", persistedCount)
	}

	t.Log("✓ HealingQueue.Remove successfully removes and persists deletion")
}

// ============================================================================
// Transaction Rollback Tests
// ============================================================================

func TestTransactionRollback_PageStore(t *testing.T) {
	db, _ := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	pageStore := NewSqlitePageStore(tx)
	pageId, err := pageStore.Add(ctx, store.PageEntry{
		PageSource: "<html>Test</html>",
		Locator:    "//button",
		B64Png:     "img",
		ContextId:  "ctx",
		ProjectId:  "proj-rollback",
		Platform:   platform.WebPlatform,
		PageType:   page.HTMLPageType,
	})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Rollback instead of commit
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify data was NOT persisted
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM page WHERE page_id = ?", pageId)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("Failed to verify rollback: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 entries after rollback, got %d", count)
	}

	t.Log("✓ Transaction rollback correctly prevents persistence")
}
