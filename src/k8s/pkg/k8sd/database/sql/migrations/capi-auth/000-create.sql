CREATE TABLE capi_auth_token (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    token TEXT NOT NULL
);
