-- Convert RFC3339 UTC back to "YYYY-MM-DD HH:MM:SS" in America/Phoenix (UTC-7).
UPDATE weigh_ins
SET created_at = strftime('%Y-%m-%d %H:%M:%S', created_at, '-7 hours')
WHERE created_at LIKE '%T%';
