INSERT INTO
    cluster_configs(key, value)
VALUES
    ("token::capi", ?)
ON CONFLICT(key) DO
    UPDATE SET value = EXCLUDED.value;

