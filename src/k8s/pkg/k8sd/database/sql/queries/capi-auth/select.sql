SELECT
    t.id
FROM
    capi_auth_token AS t
WHERE
    ( t.token = ? )
LIMIT 1
