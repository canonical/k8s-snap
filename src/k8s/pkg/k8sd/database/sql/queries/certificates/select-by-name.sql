SELECT
    key, certificate
FROM
    certificates AS c
WHERE
    ( c.name = ? )
