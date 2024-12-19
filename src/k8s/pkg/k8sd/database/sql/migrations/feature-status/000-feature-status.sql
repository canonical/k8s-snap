CREATE TABLE feature_status (
    id          INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    name        TEXT UNIQUE NOT NULL,
    message     TEXT NOT NULL,
    version     TEXT NOT NULL,
    timestamp   TEXT NOT NULL,
    enabled     BOOLEAN NOT NULL,
    UNIQUE(name)
)
-- TODO rm me