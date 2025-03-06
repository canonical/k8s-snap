CREATE TABLE helm_charts (
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    contents BLOB NOT NULL,
    PRIMARY KEY (name, version)
)
