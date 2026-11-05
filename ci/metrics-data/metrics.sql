BEGIN TRANSACTION;
INSERT INTO schema_meta (key, value) VALUES ('schema_version', '1');
INSERT INTO tests (test_id, nodeid, file, owner) VALUES (0, '', NULL, NULL);
COMMIT;
