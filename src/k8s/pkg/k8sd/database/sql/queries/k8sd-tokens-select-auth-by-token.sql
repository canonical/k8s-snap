SELECT
    username, groups
FROM
    k8sd_kubernetes_tokens AS t
WHERE
    ( t.token = ? )
