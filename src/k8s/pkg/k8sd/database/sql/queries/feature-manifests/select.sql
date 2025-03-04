SELECT
    f.name, f.version, f.manifest
FROM
    feature_manifests AS f
WHERE
    ( f.name = ? ) AND ( f.version = ? )
LIMIT 1
