CREATE TABLE weigh_ins_old (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    weight     REAL    NOT NULL,
    created_at TEXT    NOT NULL UNIQUE,
    source     TEXT,
    notes      TEXT
);
INSERT INTO weigh_ins_old (id, weight, created_at, source, notes)
    SELECT id, weight, created_at, source, notes FROM weigh_ins;
DROP TABLE weigh_ins;
ALTER TABLE weigh_ins_old RENAME TO weigh_ins;
