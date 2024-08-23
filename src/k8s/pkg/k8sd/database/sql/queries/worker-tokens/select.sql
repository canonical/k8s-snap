SELECT
    t.name, t.expiry
FROM
    worker_tokens AS t
WHERE
    ( t.token = ? )
LIMIT 1
