-- Convert "YYYY-MM-DD HH:MM:SS" (America/Phoenix, UTC-7) to RFC3339 UTC.
-- Phoenix is always UTC-7 (no DST).
UPDATE weigh_ins
SET created_at = strftime('%Y-%m-%dT%H:%M:%SZ', created_at, '+7 hours')
WHERE created_at NOT LIKE '%T%';
