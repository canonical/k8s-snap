CREATE TABLE feature_manifests (
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    manifest BLOB NOT NULL,
    PRIMARY KEY (name, version)
)
