DELETE FROM
    cluster_configs AS c
WHERE
    c.key = "worker-token::" || ?
