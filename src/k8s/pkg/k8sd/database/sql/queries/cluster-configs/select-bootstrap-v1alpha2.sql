SELECT
    c.value
FROM
    cluster_configs AS c
WHERE
    c.key = "bootstrap-v1alpha2"
