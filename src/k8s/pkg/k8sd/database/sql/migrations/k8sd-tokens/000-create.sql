CREATE TABLE k8sd_tokens (
    id          INTEGER     PRIMARY KEY AUTOINCREMENT NOT NULL,
    token       TEXT        NOT NULL,
    username    TEXT        NOT NULL,
    groups      TEXT        NOT NULL
)
