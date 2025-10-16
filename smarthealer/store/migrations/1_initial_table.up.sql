CREATE TABLE page (
    page_id INTEGER PRIMARY KEY,
    page_source TEXT NOT NULL,
    locator TEXT NOT NULL,
    b64_png TEXT,
    platform TEXT NOT NULL,
    page_type TEXT NOT NULL,
    context_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE locator (
    locator_id INTEGER PRIMARY KEY,
    page_id INTEGER NOT NULL,
    locator TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),

    FOREIGN KEY(page_id) REFERENCES page(page_id)
);