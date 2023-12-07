SELECT
    username, groups
FROM
    kubernetes_auth_tokens AS t
WHERE
    ( t.token = ? )
