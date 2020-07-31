package dbconnections

import (
	"database/sql"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/yemramirezca/http-db-service/config"
	"github.com/yemramirezca/http-db-service/db/postgres"
	"github.com/yemramirezca/http-db-service/db/repository"
	"github.com/yemramirezca/http-db-service/handler/response"
	"io/ioutil"
	"net/http"
	"fmt"
	"strings"
)


const defaultNamespace = "default"
const defaultTable = "orders"
const uri = "uri"


// InsertOrder handles an http request for creating an Order given in JSON format.
// The handler also validates the Order payload fields and handles duplicate entry or unexpected errors.
func InsertOrder(w http.ResponseWriter, r *http.Request) {
	dbURI := r.Header.Get(uri)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Error parsing request.", err)
		response.WriteCodeAndMessage(http.StatusInternalServerError, "Internal error.", w)
		return
	}

	defer r.Body.Close()
	var order repository.Order
	err = json.Unmarshal(b, &order)
	if err != nil || order.OrderId == "" || order.Total == 0 {
		response.WriteCodeAndMessage(http.StatusBadRequest, "Invalid request body, orderId / total fields cannot be empty.", w)
		return
	}
	if order.Namespace == "" {
		order.Namespace = defaultNamespace
	}
	db, err := InitDb(dbURI)
	if err != nil {
		log.Error("Error connecting db.", err)
		response.WriteCodeAndMessage(http.StatusInternalServerError, "Internal error.", w)
		return
	}
	dbRepo := &repository.OrderRepositorySQL{db, defaultTable}

	log.Debugf("Inserting order: '%+v'.", order)
	err = dbRepo.InsertOrder(order)

	switch err {
	case nil:
		w.WriteHeader(http.StatusCreated)
	case repository.ErrDuplicateKey:
		response.WriteCodeAndMessage(http.StatusConflict, fmt.Sprintf("Order %s already exists.", order.OrderId), w)
	default:
		log.Error(fmt.Sprintf("Error inserting order: '%+v'", order), err)
		response.WriteCodeAndMessage(http.StatusInternalServerError, "Internal error.", w)
	}
}

// GetOrders handles an http request for retrieving all Orders from all namespaces.
// The orders list is marshalled in JSON format and sent to the `http.ResponseWriter`
func GetOrders(w http.ResponseWriter, r *http.Request) {
	dbURI := r.Header.Get(uri)
	log.Debug("Retrieving orders")
	db, _ := InitDb(dbURI)
	dbRepo := &repository.OrderRepositorySQL{db, defaultTable}
	orders, err := dbRepo.GetOrders()
	if err != nil {
		log.Error("Error retrieving orders.", err)
		response.WriteCodeAndMessage(http.StatusInternalServerError, "Internal error.", w)
		return
	}

	if err = respondOrders(orders, w); err != nil {
		log.Error("Error sending orders response.", err)
		response.WriteCodeAndMessage(http.StatusInternalServerError, "Internal error.", w)
		return
	}
}

// GetNamespaceOrders handles an http request for retrieving all Orders from a namespace specified as a path variable.
// The orders list is marshalled in JSON format and sent to the `http.ResponseWriter`.
func GetNamespaceOrders(w http.ResponseWriter, r *http.Request) {
	dbURI := r.Header.Get(uri)
	ns, exists := mux.Vars(r)["namespace"]
	if !exists {
		response.WriteCodeAndMessage(http.StatusBadRequest, "No namespace provided.", w)
		return
	}

	log.Debugf("Retrieving orders for namespace: %s\n", ns)
	db, _ := InitDb(dbURI)
	dbRepo := &repository.OrderRepositorySQL{db, defaultTable}
	orders, err := dbRepo.GetNamespaceOrders(ns)
	if err != nil {
		log.Error("Error retrieving orders.", err)
		response.WriteCodeAndMessage(http.StatusInternalServerError, "Internal error.", w)
		return
	}

	if err = respondOrders(orders, w); err != nil {
		log.Error("Error sending orders response.", err)
		response.WriteCodeAndMessage(http.StatusInternalServerError, "Internal error.", w)
		return
	}
}

func respondOrders(orders []repository.Order, w http.ResponseWriter) error {
	body, err := json.Marshal(orders)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(body); err != nil {
		return err
	}
	return nil
}

// DeleteOrders handles an http request for deleting all Orders from all namespaces.
func DeleteOrders(w http.ResponseWriter, r *http.Request) {
	dbURI := r.Header.Get(uri)
	log.Debug("Deleting all orders")
	db, _ := InitDb(dbURI)
	dbRepo := &repository.OrderRepositorySQL{db, defaultTable}
	if err := dbRepo.DeleteOrders(); err != nil {
		log.Error("Error deleting orders.", err)
		response.WriteCodeAndMessage(http.StatusInternalServerError, "Internal error.", w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteNamespaceOrders handles an http request for deleting all Orders from a namespace specified as a path variable.
func  DeleteNamespaceOrders(w http.ResponseWriter, r *http.Request) {
	dbURI := r.Header.Get(uri)
	db, _ := InitDb(dbURI)
	dbRepo := &repository.OrderRepositorySQL{db, defaultTable}
	ns, exists := mux.Vars(r)["namespace"]
	if !exists {
		response.WriteCodeAndMessage(http.StatusBadRequest, "No namespace provided.", w)
		return
	}

	log.Debugf("Deleting orders in namespace %s\n", ns)
	if err := dbRepo.DeleteNamespaceOrders(ns); err != nil {
		log.Errorf("Deleting orders in namespace %s\n. %s", ns, err)
		response.WriteCodeAndMessage(http.StatusInternalServerError, "Internal error.", w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func InitDb(conexionString string) (*sql.DB, error) {
	db, err := sql.Open(config.PostgresDriverName, conexionString)
	if err != nil {
		return nil, errors.Wrapf(err, "while establishing connection to '%s'", config.PostgresDriverName)
	}

	log.Debug("Testing connection")
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "while testing DB connection")
	}
	q := strings.Replace(postgres.PostgresTableCreationQuery, "{name}", repository.SanitizeSQLArg("orders"), -1)
	log.Debugf("Ensuring table exists. Running query: '%q'.", q)
	if _, err := db.Exec(q); err != nil {
		return nil, errors.Wrap(err, "while initiating DB table")
	}

	return db, nil
}