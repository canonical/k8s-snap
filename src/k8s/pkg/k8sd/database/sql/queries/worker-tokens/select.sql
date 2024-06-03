SELECT
    t.name
FROM
    worker_tokens AS t
WHERE
    ( t.token = ? )
LIMIT 1
