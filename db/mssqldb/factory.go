package mssqldb

import (
	"database/sql"
	"github.com/pkg/errors"
	"github.com/yemramirezca/http-db-service/config"
	"github.com/yemramirezca/http-db-service/db/repository"
	"strings"

	log "github.com/Sirupsen/logrus"
	_ "github.com/denisenkom/go-mssqldb" //MSSQL driver initialization
)

type Mssql struct {
	DBCfg config.Config
}

func (db *Mssql) DBConnectionString() string {
	connectionURL := newSQLServerConnectionURL(
		db.DBCfg.User, db.DBCfg.Pass, db.DBCfg.Host, db.DBCfg.Name, db.DBCfg.Port)
	conexionString := strings.Replace(connectionURL.String(), connectionURL.User.String() + "@", "***:***@", 1)
	log.Debugf("Establishing connection with '%s'. Connection string: '%q'", config.SQLServerDriverName,
		conexionString)
	return connectionURL.String()
}

func (ds *Mssql) NewOrderRepositoryDb() (repository.OrderRepository, error) {
	var (
		database repository.DBQuerier
		err error
	)

	if database, err = ds.InitDb(); err != nil {
		return nil, errors.Wrap(err, "Error loading db configuration %v.")
	}
	return &repository.OrderRepositorySQL{database, ds.DBCfg  .DbOrdersTableName}, nil
}

// InitDb creates and tests a database connection using the configuration given in dbConfig.
// After it establishes a connection it also ensures that the table exists.
func (ds *Mssql)InitDb() (*sql.DB, error) {
	conexionString := ds.DBConnectionString()
	db, err := sql.Open(config.SQLServerDriverName, conexionString)
	if err != nil {
		return nil, errors.Wrapf(err, "while establishing connection to '%s'", config.SQLServerDriverName)
	}

	log.Debug("Testing connection")
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "while testing DB connection")
	}

	q := strings.Replace(sqlServerTableCreationQuery, "{name}", repository.SanitizeSQLArg(ds.DBCfg.DbOrdersTableName), -1)
	log.Debugf("Ensuring table exists. Running query: '%q'.", q)
	if _, err := db.Exec(q); err != nil {
		return nil, errors.Wrap(err, "while initiating DB table")
	}

	return db, nil
}
