CREATE TABLE worker_nodes (
  id                   INTEGER   PRIMARY  KEY    AUTOINCREMENT  NOT  NULL,
  name                 TEXT      NOT      NULL,
  UNIQUE(name)
)
