INSERT INTO
    cluster_configs(key, value)
VALUES
    ("worker-token", ?)
ON CONFLICT(key) DO
    UPDATE SET value = EXCLUDED.value;
