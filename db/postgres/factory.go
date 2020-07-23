package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/yemramirezca/http-db-service/config"
	"github.com/yemramirezca/http-db-service/db/repository"
	"strings"
)

type Postgres struct {
	DBCfg config.Config
}

func (db *Postgres) DBConnectionString() string {
	dsn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable",
		db.DBCfg.Host,
		db.DBCfg.Name,
		db.DBCfg.User,
		db.DBCfg.Pass)
	return dsn
}

func (ds *Postgres) NewOrderRepositoryDb() (repository.OrderRepository, error) {
	var (
		database repository.DBQuerier
		err error
	)

	if database, err = ds.InitDb(); err != nil {
		return nil, errors.Wrap(err, "Error loading db configuration %v.")
	}
	return &repository.OrderRepositorySQL{database, ds.DBCfg.DbOrdersTableName}, nil
}

func (ds *Postgres)InitDb() (*sql.DB, error) {
	conexionString := ds.DBConnectionString()
	db, err := sql.Open(config.PostgresDriverName, conexionString)
	if err != nil {
		return nil, errors.Wrapf(err, "while establishing connection to '%s'", config.PostgresDriverName)
	}

	log.Debug("Testing connection")
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "while testing DB connection")
	}
	q := strings.Replace(postgresTableCreationQuery, "{name}", SanitizeSQLArg(ds.DBCfg.DbOrdersTableName), -1)
	log.Debugf("Ensuring table exists. Running query: '%q'.", q)
	if _, err := db.Exec(q); err != nil {
		return nil, errors.Wrap(err, "while initiating DB table")
	}

	return db, nil
}