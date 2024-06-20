SELECT
    EXISTS (
        SELECT 1
        FROM cluster_configs AS c
        WHERE c.key = 'token::capi' AND c.value = ?
    )
