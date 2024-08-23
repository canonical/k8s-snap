ALTER TABLE worker_tokens
ADD COLUMN expiry DATETIME;

-- set default value for existing values.
UPDATE worker_tokens
SET expiry = '2100-01-01 23:59:59'
WHERE expiry IS NULL;
