package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestEnsureMigrations(t *testing.T) {
	// Create a temporary directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_migrations.db")

	// Run migrations
	err := EnsureMigrations(dbPath)
	if err != nil {
		t.Fatalf("EnsureMigrations failed: %v", err)
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatalf("Database file was not created at %s", dbPath)
	}

	// Open database to verify schema
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test 1: Verify migrations table exists
	t.Run("MigrationsTableExists", func(t *testing.T) {
		var tableName string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&tableName)
		if err != nil {
			t.Errorf("schema_migrations table does not exist: %v", err)
		}
	})

	// Test 2: Verify migration version is set
	t.Run("MigrationVersionSet", func(t *testing.T) {
		var version int
		var dirty bool
		err := db.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty)
		if err != nil {
			t.Errorf("Failed to query schema_migrations: %v", err)
		}
		if version == 0 {
			t.Error("Migration version is 0, expected a positive value")
		}
		if dirty {
			t.Error("Migration is marked as dirty, expected clean state")
		}
		t.Logf("Current migration version: %d, dirty: %v", version, dirty)
	})

	// Test 3: Verify 'page' table exists and has correct schema
	t.Run("PageTableExists", func(t *testing.T) {
		var tableName string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='page'").Scan(&tableName)
		if err != nil {
			t.Fatalf("page table does not exist: %v", err)
		}

		// Verify columns
		expectedColumns := []string{
			"page_id", "page_source", "locator", "b64_png",
			"platform", "page_type", "context_id", "project_id", "created_at",
		}

		for _, col := range expectedColumns {
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('page') WHERE name = ?", col).Scan(&count)
			if err != nil {
				t.Errorf("Failed to check column %s: %v", col, err)
			}
			if count == 0 {
				t.Errorf("Column %s does not exist in page table", col)
			}
		}
	})

	// Test 4: Verify 'locator' table exists and has correct schema
	t.Run("LocatorTableExists", func(t *testing.T) {
		var tableName string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='locator'").Scan(&tableName)
		if err != nil {
			t.Fatalf("locator table does not exist: %v", err)
		}

		// Verify columns
		expectedColumns := []string{
			"locator_id", "page_id", "locator", "description", "created_at",
		}

		for _, col := range expectedColumns {
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('locator') WHERE name = ?", col).Scan(&count)
			if err != nil {
				t.Errorf("Failed to check column %s: %v", col, err)
			}
			if count == 0 {
				t.Errorf("Column %s does not exist in locator table", col)
			}
		}
	})

	// Test 5: Verify 'description_queue' table exists
	t.Run("DescriptionQueueTableExists", func(t *testing.T) {
		var tableName string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='description_queue'").Scan(&tableName)
		if err != nil {
			t.Fatalf("description_queue table does not exist: %v", err)
		}

		// Verify columns
		expectedColumns := []string{"page_id", "locator_id", "created_at"}

		for _, col := range expectedColumns {
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('description_queue') WHERE name = ?", col).Scan(&count)
			if err != nil {
				t.Errorf("Failed to check column %s: %v", col, err)
			}
			if count == 0 {
				t.Errorf("Column %s does not exist in description_queue table", col)
			}
		}
	})

	// Test 6: Verify 'healing_queue' table exists
	t.Run("HealingQueueTableExists", func(t *testing.T) {
		var tableName string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='healing_queue'").Scan(&tableName)
		if err != nil {
			t.Fatalf("healing_queue table does not exist: %v", err)
		}

		// Verify columns
		expectedColumns := []string{"id", "info_json", "opt_json", "created_at"}

		for _, col := range expectedColumns {
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('healing_queue') WHERE name = ?", col).Scan(&count)
			if err != nil {
				t.Errorf("Failed to check column %s: %v", col, err)
			}
			if count == 0 {
				t.Errorf("Column %s does not exist in healing_queue table", col)
			}
		}
	})

	// Test 7: Test inserting data into tables
	t.Run("InsertAndQueryData", func(t *testing.T) {
		// Insert into page table
		_, err := db.Exec(`
			INSERT INTO page (page_source, locator, b64_png, platform, page_type, context_id, project_id)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, "<html>test</html>", "//button[@id='test']", "base64data", "Android", "HTML", "ctx-123", "proj-123")
		if err != nil {
			t.Errorf("Failed to insert into page table: %v", err)
		}

		// Query back
		var pageID int
		err = db.QueryRow("SELECT page_id FROM page WHERE context_id = ?", "ctx-123").Scan(&pageID)
		if err != nil {
			t.Errorf("Failed to query page table: %v", err)
		}

		if pageID == 0 {
			t.Error("Expected page_id to be non-zero")
		}

		// Insert into locator table
		_, err = db.Exec(`
			INSERT INTO locator (page_id, locator, description)
			VALUES (?, ?, ?)
		`, pageID, "//button[@id='test']", "Test button")
		if err != nil {
			t.Errorf("Failed to insert into locator table: %v", err)
		}

		// Insert into description_queue
		_, err = db.Exec(`
			INSERT INTO description_queue (page_id, locator_id)
			VALUES (?, ?)
		`, pageID, 1)
		if err != nil {
			t.Errorf("Failed to insert into description_queue: %v", err)
		}

		// Insert into healing_queue
		_, err = db.Exec(`
			INSERT INTO healing_queue (info_json, opt_json)
			VALUES (?, ?)
		`, `{"test": "info"}`, `{"test": "opt"}`)
		if err != nil {
			t.Errorf("Failed to insert into healing_queue: %v", err)
		}
	})
}

func TestEnsureMigrationsIdempotent(t *testing.T) {
	// Create a temporary directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_idempotent.db")

	// Run migrations first time
	err := EnsureMigrations(dbPath)
	if err != nil {
		t.Fatalf("First EnsureMigrations failed: %v", err)
	}

	// Run migrations second time (should be idempotent)
	err = EnsureMigrations(dbPath)
	if err != nil {
		t.Fatalf("Second EnsureMigrations failed: %v", err)
	}

	// Verify database is still functional
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	var version int
	err = db.QueryRow("SELECT version FROM schema_migrations").Scan(&version)
	if err != nil {
		t.Errorf("Failed to query schema_migrations after idempotent test: %v", err)
	}
}

func TestEnsureMigrationsWithInvalidPath(t *testing.T) {
	// Test with invalid path (should fail gracefully)
	invalidPath := "/invalid/path/that/does/not/exist/test.db"

	err := EnsureMigrations(invalidPath)
	if err == nil {
		t.Error("Expected error with invalid path, but got nil")
	}
}

func TestEnsureMigrationsEmptyPath(t *testing.T) {
	// Test with empty path (should use default or fail)
	err := EnsureMigrations("")
	// This should either work with a default path or fail with a specific error
	// The behavior depends on your implementation
	t.Logf("Empty path result: %v", err)
}
