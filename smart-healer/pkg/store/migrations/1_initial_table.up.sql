CREATE TABLE page (
    page_id INTEGER PRIMARY KEY,
    page_source TEXT NOT NULL,
    b64_png TEXT,
    platform TEXT NOT NULL,
    platform_type TEXT NOT NULL,
    context_id TEXT,
    project_id TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE locator (
    locator_id INTEGER PRIMARY KEY,
    page_id INTEGER NOT NULL,
    locator TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TEXT NOT NULL

    FOREIGN KEY(page_id) REFERECES page(page_id)
);