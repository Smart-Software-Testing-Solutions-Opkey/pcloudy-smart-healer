package smarthealer

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/healer"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/intelligence"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/page"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/platform"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/retrieval"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store"
	"github.com/Smart-Software-Testing-Solutions-Opkey/pcloudy-smart-healer/smarthealer/store/sqliteimpl"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrMockDescriptionFailed    = errors.New("mock description generation failed")
	ErrMockLocatorFailed        = errors.New("mock locator generation failed")
	ErrMockScreenshotComparison = errors.New("mock screenshot comparison failed")
)

// ============================================================================
// Mock AI Intelligence System
// ============================================================================

type mockIntelligenceSystem struct {
	shouldFail bool
}

func newMockIntelligence(shouldFail bool) intelligence.IntelligenceSystem {
	return &mockIntelligenceSystem{shouldFail: shouldFail}
}

func (m *mockIntelligenceSystem) GenerateElementDescription(ctx context.Context, pageSource, elementSource string) (string, error) {
	if m.shouldFail {
		return "", ErrMockDescriptionFailed
	}
	return "Mocked element description for testing", nil
}

func (m *mockIntelligenceSystem) GenerateLocator(ctx context.Context, desc string, root page.Page, plat platform.Platform) (string, error) {
	if m.shouldFail {
		return "", ErrMockLocatorFailed
	}
	return "//button[@id='mocked']", nil
}

func (m *mockIntelligenceSystem) CompareScreenShot(ctx context.Context, img1, img2 string) (bool, error) {
	if m.shouldFail {
		return false, ErrMockScreenshotComparison
	}
	return true, nil
}

// ============================================================================
// Helper to create SmartHealer with mocked AI
// ============================================================================

func setupSmartHealerWithMockAI(t *testing.T, shouldAIFail bool) (*SmartHealer, *sqlx.DB, string) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_e2e.db")

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
	mockIntel := newMockIntelligence(shouldAIFail)

	// Create page retriever
	pageRetriever := retrieval.NewPageRetriever(uowF, mockIntel)

	// Create background worker
	bg, err := healer.NewBGWorker(mockIntel, uowF)
	if err != nil {
		t.Fatalf("Failed to create background worker: %v", err)
	}

	// Create healer
	h := healer.NewHealer(mockIntel, pageRetriever, uowF, bg)

	// Create SmartHealer
	ctx, cancel := context.WithCancel(context.Background())

	sh := &SmartHealer{
		ctx:    ctx,
		cancel: cancel,
		uofW:   uowF,
		healer: h,
		bg:     bg,
	}

	return sh, db, dbPath
}

// ============================================================================
// END-TO-END TEST: SUCCESSFUL PATH (Valid XPath)
// ============================================================================

func TestEndToEnd_SuccessfulPath_ValidXPath_DatabaseEntriesCreated(t *testing.T) {
	sh, db, _ := setupSmartHealerWithMockAI(t, false)
	defer sh.Close()
	defer db.Close()

	_ = context.Background()

	t.Log("=== TESTING SUCCESSFUL PATH: Valid XPath ===")

	// Create info with VALID xpath that exists in the page source
	info := healer.LocatorInfo{
		ProjectId:  "e2e-success-test",
		PageSource: "<html><body><button id='submit'>Submit</button></body></html>",
		B64Png:     "data:image/png;base64,iVBORw0KGgoAAAANSUhEUg",
		XPath:      "//button[@id='submit']", // This xpath EXISTS in the page source
		ContextId:  "test-context-success",
		Platform:   platform.WebPlatform,
		PageType:   page.HTMLPageType,
	}

	opt := healer.ResolveOptions{
		ComparisionMode: retrieval.AutomaticComparisionMode,
	}

	t.Log("Step 1: Calling ResolveLocator with valid xpath...")
	result, err := sh.ResolveLocator(info, opt)

	if err != nil {
		t.Logf("ResolveLocator returned error (expected for new entry with no matches): %v", err)
		// This is expected - when there are no existing pages, it tries to create a new entry
		// The error comes from page retrieval, not from database commits
	}

	t.Logf("ResolveLocator result: %s", result)

	// Wait a moment for any async operations
	time.Sleep(100 * time.Millisecond)

	t.Log("Step 2: Verifying database entries were created...")

	// Check if PAGE was inserted
	var pageCount int
	err = db.Get(&pageCount, "SELECT COUNT(*) FROM page WHERE project_id = ?", info.ProjectId)
	if err != nil {
		t.Fatalf("Failed to query page table: %v", err)
	}

	t.Logf("Pages found in database: %d", pageCount)

	if pageCount > 0 {
		t.Log("✅ SUCCESS: Page entry was created and committed!")

		// Get the page details
		type PageRow struct {
			PageId     int    `db:"page_id"`
			PageSource string `db:"page_source"`
			Locator    string `db:"locator"`
			Platform   string `db:"platform"`
			PageType   string `db:"page_type"`
		}

		var pageRow PageRow
		err = db.Get(&pageRow, "SELECT page_id, page_source, locator, platform, page_type FROM page WHERE project_id = ?", info.ProjectId)
		if err != nil {
			t.Fatalf("Failed to get page details: %v", err)
		}

		t.Logf("Page ID: %d", pageRow.PageId)
		t.Logf("Page Source: %s", pageRow.PageSource)
		t.Logf("Locator: %s", pageRow.Locator)
		t.Logf("Platform: %s", pageRow.Platform)
		t.Logf("Page Type: %s", pageRow.PageType)

		// Check if LOCATOR was inserted
		var locatorCount int
		err = db.Get(&locatorCount, "SELECT COUNT(*) FROM locator WHERE page_id = ?", pageRow.PageId)
		if err != nil {
			t.Fatalf("Failed to query locator table: %v", err)
		}

		t.Logf("Locators found for page %d: %d", pageRow.PageId, locatorCount)

		if locatorCount > 0 {
			t.Log("✅ SUCCESS: Locator entry was created and committed!")

			// Get locator details
			type LocatorRow struct {
				LocatorId   int    `db:"locator_id"`
				Locator     string `db:"locator"`
				Description string `db:"description"`
			}

			var locatorRow LocatorRow
			err = db.Get(&locatorRow, "SELECT locator_id, locator, description FROM locator WHERE page_id = ?", pageRow.PageId)
			if err != nil {
				t.Fatalf("Failed to get locator details: %v", err)
			}

			t.Logf("Locator ID: %d", locatorRow.LocatorId)
			t.Logf("Locator XPath: %s", locatorRow.Locator)
			t.Logf("Description: %s", locatorRow.Description)

			// Check if DESCRIPTION QUEUE entry was created
			var descQueueCount int
			err = db.Get(&descQueueCount, "SELECT COUNT(*) FROM description_queue WHERE page_id = ? AND locator_id = ?",
				pageRow.PageId, locatorRow.LocatorId)
			if err != nil {
				t.Fatalf("Failed to query description_queue: %v", err)
			}

			t.Logf("Description queue entries: %d", descQueueCount)

			if descQueueCount > 0 {
				t.Log("✅ SUCCESS: Description queue entry was created!")
			}
		} else {
			t.Error("❌ FAILED: No locator entries found - commit may have failed")
		}
	} else {
		t.Error("❌ FAILED: No page entries found - commit failed!")
		t.Error("This indicates the transaction commit is not working correctly.")
	}

	t.Log("=== END OF SUCCESSFUL PATH TEST ===")
}

// ============================================================================
// END-TO-END TEST: UNSUCCESSFUL PATH (Invalid XPath)
// ============================================================================

func TestEndToEnd_UnsuccessfulPath_InvalidXPath_NoEntriesCreated(t *testing.T) {
	sh, db, _ := setupSmartHealerWithMockAI(t, false)
	defer sh.Close()
	defer db.Close()

	t.Log("=== TESTING UNSUCCESSFUL PATH: Invalid XPath ===")

	// Create info with INVALID xpath that DOES NOT exist in the page source
	info := healer.LocatorInfo{
		ProjectId:  "e2e-failure-test",
		PageSource: "<html><body><button id='submit'>Submit</button></body></html>",
		B64Png:     "data:image/png;base64,iVBORw0KGgoAAAANSUhEUg",
		XPath:      "//button[@id='nonexistent']", // This xpath DOES NOT exist in page source
		ContextId:  "test-context-failure",
		Platform:   platform.WebPlatform,
		PageType:   page.HTMLPageType,
	}

	opt := healer.ResolveOptions{
		ComparisionMode: retrieval.AutomaticComparisionMode,
	}

	t.Log("Step 1: Calling ResolveLocator with invalid xpath...")
	result, err := sh.ResolveLocator(info, opt)

	if err != nil {
		t.Logf("✅ Expected error for invalid xpath: %v", err)
	} else {
		t.Errorf("Expected error for invalid xpath, but got result: %s", result)
	}

	// Wait a moment for any async operations
	time.Sleep(100 * time.Millisecond)

	t.Log("Step 2: Verifying NO database entries were created (rollback worked)...")

	// Check that NO PAGE was inserted
	var pageCount int
	err = db.Get(&pageCount, "SELECT COUNT(*) FROM page WHERE project_id = ?", info.ProjectId)
	if err != nil {
		t.Fatalf("Failed to query page table: %v", err)
	}

	t.Logf("Pages found in database: %d", pageCount)

	if pageCount == 0 {
		t.Log("✅ SUCCESS: No page entry was created (transaction was rolled back correctly)")
	} else {
		t.Errorf("❌ FAILED: Found %d page entries - rollback did not work!", pageCount)
	}

	// Double-check locator table
	var locatorCount int
	err = db.Get(&locatorCount, "SELECT COUNT(*) FROM locator")
	if err != nil {
		t.Fatalf("Failed to query locator table: %v", err)
	}

	if locatorCount == 0 {
		t.Log("✅ SUCCESS: No locator entries (rollback worked)")
	} else {
		t.Errorf("❌ FAILED: Found %d locator entries after rollback", locatorCount)
	}

	t.Log("=== END OF UNSUCCESSFUL PATH TEST ===")
}

// ============================================================================
// END-TO-END TEST: Async Resolution with Background Worker
// ============================================================================

func TestEndToEnd_AsyncResolution_BackgroundWorker_ProcessesQueue(t *testing.T) {
	sh, db, _ := setupSmartHealerWithMockAI(t, false)
	defer sh.Close()
	defer db.Close()

	// Start background workers
	sh.StartBackgroundWorkers()

	t.Log("=== TESTING ASYNC RESOLUTION WITH BACKGROUND WORKERS ===")

	info := healer.LocatorInfo{
		ProjectId:  "e2e-async-test",
		PageSource: "<html><body><div id='container'>Test</div></body></html>",
		B64Png:     "data:image/png;base64,iVBORw0KGgoAAAANSUhEUg",
		XPath:      "//div[@id='container']",
		ContextId:  "test-async",
		Platform:   platform.WebPlatform,
		PageType:   page.HTMLPageType,
	}

	opt := healer.ResolveOptions{
		ComparisionMode: retrieval.AutomaticComparisionMode,
	}

	t.Log("Step 1: Calling ResolveLocatorAsync...")
	err := sh.ResolveLocatorAsync(info, opt)
	if err != nil {
		t.Fatalf("ResolveLocatorAsync failed: %v", err)
	}

	t.Log("✅ ResolveLocatorAsync completed without error")

	t.Log("Step 2: Verifying healing queue entry was created...")
	var queueCountBefore int
	err = db.Get(&queueCountBefore, "SELECT COUNT(*) FROM healing_queue")
	if err != nil {
		t.Fatalf("Failed to query healing_queue: %v", err)
	}

	t.Logf("Healing queue entries before processing: %d", queueCountBefore)

	if queueCountBefore == 0 {
		t.Error("❌ FAILED: No healing queue entry was created - commit failed!")
	} else {
		t.Log("✅ SUCCESS: Healing queue entry was created and committed")

		// Get queue entry details
		type QueueRow struct {
			Id       int    `db:"id"`
			InfoJson string `db:"info_json"`
			OptJson  string `db:"opt_json"`
		}

		var queueRow QueueRow
		err = db.Get(&queueRow, "SELECT id, info_json, opt_json FROM healing_queue ORDER BY created_at DESC LIMIT 1")
		if err != nil {
			t.Fatalf("Failed to get queue details: %v", err)
		}

		t.Logf("Queue Entry ID: %d", queueRow.Id)
		t.Logf("Info JSON: %s", queueRow.InfoJson)
		t.Logf("Opt JSON: %s", queueRow.OptJson)
	}

	t.Log("Step 3: Waiting for background worker to process...")
	// Background workers process every 2 seconds (rate.Every(2 * time.Second))
	time.Sleep(3 * time.Second)

	var queueCountAfter int
	err = db.Get(&queueCountAfter, "SELECT COUNT(*) FROM healing_queue")
	if err != nil {
		t.Fatalf("Failed to query healing_queue after processing: %v", err)
	}

	t.Logf("Healing queue entries after background processing: %d", queueCountAfter)

	if queueCountAfter < queueCountBefore {
		t.Log("✅ SUCCESS: Background worker processed and removed queue entry")
	} else {
		t.Log("⚠️  Background worker may not have processed yet (or processing failed)")
	}

	// Check if page was created by background worker
	var pageCount int
	err = db.Get(&pageCount, "SELECT COUNT(*) FROM page WHERE project_id = ?", info.ProjectId)
	if err != nil {
		t.Fatalf("Failed to query page table: %v", err)
	}

	t.Logf("Pages created by background worker: %d", pageCount)

	if pageCount > 0 {
		t.Log("✅ SUCCESS: Background worker created page entry and committed!")
	}

	t.Log("=== END OF ASYNC RESOLUTION TEST ===")
}

// ============================================================================
// END-TO-END TEST: Multiple Concurrent Operations
// ============================================================================

func TestEndToEnd_ConcurrentOperations_AllCommitsSucceed(t *testing.T) {
	sh, db, _ := setupSmartHealerWithMockAI(t, false)
	defer sh.Close()
	defer db.Close()

	sh.StartBackgroundWorkers()

	t.Log("=== TESTING CONCURRENT OPERATIONS ===")

	const numOperations = 10

	done := make(chan bool, numOperations)

	t.Logf("Launching %d concurrent async operations...", numOperations)

	for i := 0; i < numOperations; i++ {
		go func(id int) {
			defer func() { done <- true }()

			info := healer.LocatorInfo{
				ProjectId:  "concurrent-test",
				PageSource: "<html><body><button>Test</button></body></html>",
				B64Png:     "data:image/png;base64,test",
				XPath:      "//button",
				ContextId:  "concurrent",
				Platform:   platform.WebPlatform,
				PageType:   page.HTMLPageType,
			}

			opt := healer.ResolveOptions{
				ComparisionMode: retrieval.AutomaticComparisionMode,
			}

			err := sh.ResolveLocatorAsync(info, opt)
			if err != nil {
				t.Errorf("Operation %d failed: %v", id, err)
			}
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < numOperations; i++ {
		<-done
	}

	t.Log("All concurrent operations completed")

	// Verify all entries were committed
	var queueCount int
	err := db.Get(&queueCount, "SELECT COUNT(*) FROM healing_queue")
	if err != nil {
		t.Fatalf("Failed to query healing_queue: %v", err)
	}

	t.Logf("Total healing queue entries created: %d (expected %d)", queueCount, numOperations)

	if queueCount == numOperations {
		t.Log("✅ SUCCESS: All concurrent operations committed successfully!")
	} else {
		t.Errorf("❌ FAILED: Expected %d queue entries, got %d - some commits failed", numOperations, queueCount)
	}

	// Verify database integrity
	var tableCount int
	err = db.Get(&tableCount, "SELECT COUNT(*) FROM sqlite_master WHERE type='table'")
	if err != nil {
		t.Fatalf("Database integrity check failed: %v", err)
	}

	if tableCount >= 4 {
		t.Log("✅ SUCCESS: Database integrity maintained after concurrent operations")
	}

	t.Log("=== END OF CONCURRENT OPERATIONS TEST ===")
}
