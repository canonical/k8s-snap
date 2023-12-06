SELECT
    token
FROM
    k8sd_kubernetes_tokens AS t
WHERE
    ( t.username = ? AND t.groups = ? )
