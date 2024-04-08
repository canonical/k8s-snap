INSERT INTO
    cluster_configs(key, value)
VALUES
    ("v1alpha2", ?)
ON CONFLICT(key) DO
    UPDATE SET value = EXCLUDED.value;
