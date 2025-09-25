CREATE TABLE description_queue (
    page_id INTEGER NOT NULL,
    locator_id INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT(CURRENT_TIMESTAMP),
    
    PRIMARY KEY (page_id, locator_id)
);
    
CREATE TABLE healing_queue (
    id INTEGER PRIMARY KEY,
    info_json TEXT NOT NULL,
    opt_json TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT(CURRENT_TIMESTAMP)
);