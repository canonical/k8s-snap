INSERT INTO
    feature_status(name, message, version, timestamp, enabled)
VALUES
    (?, ?, ?, ?, ?)
ON CONFLICT(name) DO UPDATE SET
    message=excluded.message,
    version=excluded.version,
    timestamp=excluded.timestamp,
    enabled=excluded.enabled;
