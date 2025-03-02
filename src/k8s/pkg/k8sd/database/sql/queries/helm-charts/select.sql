SELECT
    c.name, c.version, c.contents
FROM
    helm_charts AS c
WHERE
    ( c.name = ? ) AND ( c.version = ? )
LIMIT 1
