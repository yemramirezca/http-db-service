package mssqldb

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"github.com/yemramirezca/http-db-service/config"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
)

func TestFactory(t *testing.T) {
	if os.Getenv("Host") == "" {
		t.Skip("skipping test; DB Config not set")
	}

	var dbCfg config.Config
	assert.NoError(t, envconfig.Init(&dbCfg))
	mssql := Mssql{dbCfg}
	// when initiating
	db, err := mssql.InitDb()
	require.NoError(t, err)
	defer db.Close()

	// then
	rows, err := db.Query("SELECT 1 FROM sysobjects WHERE xtype = 'U' AND name = ?", dbCfg.DbOrdersTableName)
	require.NoError(t, err)
	defer rows.Close()
	// check that table exists
	assert.True(t, rows.Next())

	//cleanup
	dropTable(db, dbCfg.DbOrdersTableName)
}

func dropTable(db *sql.DB, tableName string) {
	db.Exec(fmt.Sprintf("DROP TABLE %s", tableName))
}
