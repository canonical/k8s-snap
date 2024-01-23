INSERT INTO cluster_configs (key, value)
VALUES (?, ?)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;
