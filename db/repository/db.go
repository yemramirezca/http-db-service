package repository

import (
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/yemramirezca/http-db-service/config"
	"io"
	"regexp"
	"strings"
)

const (
	insertQuery         = "INSERT INTO %s (order_id, namespace, total) VALUES ($1, $2, $3)"
	getQuery            = "SELECT * FROM %s"
	getNSQuery          = "SELECT * FROM %s WHERE namespace = ?"
	deleteQuery         = "DELETE FROM %s"
	deleteNSQuery       = "DELETE FROM %s WHERE namespace = ?"
	PrimaryKeyViolation = 2627
	DefaultTable        = "orders"
	TableCreationQuery  = `
    CREATE TABLE IF NOT EXISTS {name} (
      order_id VARCHAR(64),
      namespace VARCHAR(64),
      total DECIMAL(8,2),
      PRIMARY KEY (order_id, namespace)
    )
`
)

type Database interface {
	DBConnectionString() string
	NewOrderRepositoryDb() (OrderRepository, error)
}

type OrderRepositorySQL struct {
	Database        DBQuerier
	OrdersTableName string
}

//go:generate mockery -name DBQuerier -inpkg
type DBQuerier interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	io.Closer
}


type sqlError interface {
	sqlErrorNumber() int32
}

func (repository *OrderRepositorySQL) InsertOrder(order Order) error {
	q := fmt.Sprintf(insertQuery, SanitizeSQLArg(repository.OrdersTableName))
	log.Debugf("Running insert order query: '%q'.", q)
	_, err := repository.Database.Exec(q, order.OrderId, order.Namespace, order.Total)


	if errorWithNumber, ok := err.(sqlError); ok {
		if errorWithNumber.sqlErrorNumber() == PrimaryKeyViolation {
			return ErrDuplicateKey
		}
	}

	return errors.Wrap(err, "while inserting order")
}

func (repository *OrderRepositorySQL) GetOrders() ([]Order, error) {
	q := fmt.Sprintf(getQuery, SanitizeSQLArg(repository.OrdersTableName))
	log.Debugf("Quering orders: '%q'.", q)
	rows, err := repository.Database.Query(q)

	if err != nil {
		return nil, errors.Wrap(err, "while reading orders from DB")
	}

	defer rows.Close()
	return readFromResult(rows)
}

func (repository *OrderRepositorySQL) GetNamespaceOrders(ns string) ([]Order, error) {
	q := fmt.Sprintf(getNSQuery, SanitizeSQLArg(repository.OrdersTableName))
	log.Debugf("Quering orders for namespace: '%q'.", q)
	rows, err := repository.Database.Query(q, ns)

	if err != nil {
		return nil, errors.Wrapf(err, "while reading orders for namespace: '%q' from DB", ns)
	}

	defer rows.Close()
	return readFromResult(rows)
}

func (repository *OrderRepositorySQL) DeleteOrders() error {
	q := fmt.Sprintf(deleteQuery, SanitizeSQLArg(repository.OrdersTableName))
	log.Debugf("Deleting orders: '%q'.", q)
	_, err := repository.Database.Exec(q)

	if err != nil {
		return errors.Wrap(err, "while deleting orders")
	}
	return nil
}

func (repository *OrderRepositorySQL) DeleteNamespaceOrders(ns string) error {
	q := fmt.Sprintf(deleteNSQuery, SanitizeSQLArg(repository.OrdersTableName))
	log.Debugf("Deleting orders: '%q'.", q)
	_, err := repository.Database.Exec(q, ns)

	if err != nil {
		return errors.Wrap(err, "while deleting orders")
	}
	return nil
}

func readFromResult(rows *sql.Rows) ([]Order, error) {
	orderList := make([]Order, 0)
	for rows.Next() {
		order := Order{}
		if err := rows.Scan(&order.OrderId, &order.Namespace, &order.Total); err != nil {
			return []Order{}, err
		}
		orderList = append(orderList, order)
	}
	return orderList, nil
}

func (repository *OrderRepositorySQL) CleanUp() error {
	log.Debug("Removing DB table")

	if _, err := repository.Database.Exec("DROP TABLE " + SanitizeSQLArg(repository.OrdersTableName)); err != nil {
		return errors.Wrap(err, "while removing the DB table.")
	}
	if err := repository.Database.Close(); err != nil {
		return errors.Wrap(err, "while closing connection to the DB.")
	}
	return nil
}

var safeSQLRegex = regexp.MustCompile(`[^a-zA-Z0-9\.\-_]`)

// SanitizeSQLArg returns the input string sanitized for safe use in an SQL query as argument.
func SanitizeSQLArg(s string) string {
	return safeSQLRegex.ReplaceAllString(s, "")
}

func InitDb(conexionString string) (*sql.DB, error) {
	db, err := sql.Open(config.PostgresDriverName, conexionString)
	if err != nil {
		return nil, errors.Wrapf(err, "while establishing connection to '%s'", config.PostgresDriverName)
	}
	db.Stats()
	log.Debug("Testing connection")
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "while testing DB connection")
	}
	q := strings.Replace(TableCreationQuery, "{name}", SanitizeSQLArg(DefaultTable), -1)
	log.Debugf("Ensuring table exists. Running query: '%q'.", q)
	if _, err := db.Exec(q); err != nil {
		return nil, errors.Wrap(err, "while initiating DB table")
	}
	return db, nil
}