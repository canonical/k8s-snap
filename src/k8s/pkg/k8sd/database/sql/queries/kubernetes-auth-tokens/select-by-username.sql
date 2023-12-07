SELECT
    token
FROM
    kubernetes_auth_tokens AS t
WHERE
    ( t.username = ? AND t.groups = ? )
