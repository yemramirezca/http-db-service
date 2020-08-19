package handler

import (
	"encoding/json"
	"fmt"
	"github.com/yemramirezca/http-db-service/config"
	"github.com/yemramirezca/http-db-service/handler/response"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"github.com/yemramirezca/http-db-service/db/repository"
)

const defaultNamespace = "default"
const header = "user"
const user = "jason"

// Order is used to expose the Order service's basic operations using the HTTP route handler methods which extend it.
type Order struct {
	repository1 repository.OrderRepositorySQL
	repository2 repository.OrderRepositorySQL
	serviceConfig config.Service
}

// NewOrderHandler creates a new 'OrderHandler' which provides route handlers for the given OrderRepository's operations.
func NewOrderHandler(cfg config.Service) Order {
	db1,err := repository.InitDb(cfg.DBConnection1)
	if err != nil {
		log.Fatal("Unable to connect to db 1", err)
	}
	db2,err := repository.InitDb(cfg.DBConnection2)
	if err != nil {
		log.Fatal("Unable to connect to db 2", err)
	}
	return Order{repository.OrderRepositorySQL{Database:db1, OrdersTableName:repository.DefaultTable},
		repository.OrderRepositorySQL{Database:db2, OrdersTableName:repository.DefaultTable},
		cfg}
}

// InsertOrder handles an http request for creating an Order given in JSON format.
// The handler also validates the Order payload fields and handles duplicate entry or unexpected errors.
func (orderHandler Order) InsertOrder(w http.ResponseWriter, r *http.Request) {
	headerVal := r.Header.Get(header)

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


	log.Printf("Inserting order: '%+v'.", order)
	repo,err := orderHandler.getRepository(headerVal)
	if err != nil {
		response.WriteCodeAndMessage(http.StatusUnauthorized, err.Error(), w)
		return
	}
	err = repo.InsertOrder(order)

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
func (orderHandler Order) GetOrders(w http.ResponseWriter, r *http.Request) {
	headerVal := r.Header.Get(header)
	log.Printf("Retrieving orders")
	repo,err := orderHandler.getRepository(headerVal)
	if err != nil {
		response.WriteCodeAndMessage(http.StatusUnauthorized, err.Error(), w)
		return
	}
	orders, err := repo.GetOrders()

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
func (orderHandler Order) GetNamespaceOrders(w http.ResponseWriter, r *http.Request) {
	headerVal := r.Header.Get(header)
	ns, exists := mux.Vars(r)["namespace"]
	if !exists {
		response.WriteCodeAndMessage(http.StatusBadRequest, "No namespace provided.", w)
		return
	}

	log.Printf("Retrieving orders for namespace: %s\n", ns)
	repo,err := orderHandler.getRepository(headerVal)
	if err != nil {
		response.WriteCodeAndMessage(http.StatusUnauthorized, err.Error(), w)
		return
	}
	orders, err := repo.GetNamespaceOrders(ns)
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
func (orderHandler Order) DeleteOrders(w http.ResponseWriter, r *http.Request) {
	headerVal := r.Header.Get(header)
	log.Printf("Deleting all orders")
	repo,err := orderHandler.getRepository(headerVal)
	if err != nil {
		response.WriteCodeAndMessage(http.StatusUnauthorized, err.Error(), w)
		return
	}

	if err := repo.DeleteOrders(); err != nil {
		log.Error("Error deleting orders.", err)
		response.WriteCodeAndMessage(http.StatusInternalServerError, "Internal error.", w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteNamespaceOrders handles an http request for deleting all Orders from a namespace specified as a path variable.
func (orderHandler Order) DeleteNamespaceOrders(w http.ResponseWriter, r *http.Request) {
	headerVal := r.Header.Get(header)
	ns, exists := mux.Vars(r)["namespace"]
	if !exists {
		response.WriteCodeAndMessage(http.StatusBadRequest, "No namespace provided.", w)
		return
	}
	repo,err := orderHandler.getRepository(headerVal)
	if err != nil {
		response.WriteCodeAndMessage(http.StatusUnauthorized, err.Error(), w)
		return
	}
	log.Printf("Deleting orders in namespace %s\n", ns)
	if err := repo.DeleteNamespaceOrders(ns); err != nil {
		log.Errorf("Deleting orders in namespace %s\n. %s", ns, err)
		response.WriteCodeAndMessage(http.StatusInternalServerError, "Internal error.", w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}


func (orderHandler Order) getRepository(header string) (*repository.OrderRepositorySQL, error) {
	if header == user {
		return &orderHandler.repository1, nil
	}
	return &orderHandler.repository2, nil
}
