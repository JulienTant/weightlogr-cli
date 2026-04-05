CREATE TABLE IF NOT EXISTS weigh_ins (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    weight     REAL    NOT NULL,
    created_at TEXT    NOT NULL UNIQUE,
    source     TEXT,
    notes      TEXT
);
