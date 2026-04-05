-- 1. Add columns as nullable
ALTER TABLE weigh_ins ADD COLUMN updated_at TEXT;
ALTER TABLE weigh_ins ADD COLUMN deleted_at TEXT;

-- 2. Backfill updated_at from created_at
UPDATE weigh_ins SET updated_at = created_at WHERE updated_at IS NULL;

-- 3. Recreate table with updated_at NOT NULL, deleted_at nullable
CREATE TABLE weigh_ins_new (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    weight     REAL    NOT NULL,
    created_at TEXT    NOT NULL UNIQUE,
    source     TEXT,
    notes      TEXT,
    updated_at TEXT    NOT NULL,
    deleted_at TEXT
);
INSERT INTO weigh_ins_new (id, weight, created_at, source, notes, updated_at, deleted_at)
    SELECT id, weight, created_at, source, notes, updated_at, deleted_at FROM weigh_ins;
DROP TABLE weigh_ins;
ALTER TABLE weigh_ins_new RENAME TO weigh_ins;
