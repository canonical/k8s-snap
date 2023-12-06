CREATE TABLE kubernetes_auth_tokens (
    id          INTEGER     PRIMARY KEY AUTOINCREMENT NOT NULL,
    token       TEXT        NOT NULL,
    username    TEXT        NOT NULL,
    groups      TEXT        NOT NULL
)
