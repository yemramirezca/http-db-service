package mssqldb

const (

  sqlServerTableCreationQuery = `IF NOT EXISTS(SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = '{name}')
BEGIN
    CREATE TABLE {name} (
      order_id VARCHAR(64),
      namespace VARCHAR(64),
      total DECIMAL(8,2),
      PRIMARY KEY (order_id, namespace)
    )
END`
)
