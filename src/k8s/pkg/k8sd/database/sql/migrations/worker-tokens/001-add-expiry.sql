ALTER TABLE worker_tokens
ADD COLUMN expiry DATETIME DEFAULT '2100-01-01 23:59:59';
