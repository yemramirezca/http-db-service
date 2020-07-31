package main

import (
	"github.com/yemramirezca/http-db-service/handler/events"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/vrischmann/envconfig"

	_ "github.com/lib/pq"
	"github.com/yemramirezca/http-db-service/config"
	"github.com/yemramirezca/http-db-service/handler"
	r "github.com/yemramirezca/http-db-service/handler/dbconnections"
)

func main() {
	log.Println("Starting service...")

	var cfg config.Service
	if err := envconfig.Init(&cfg); err != nil {
		log.Panicf("Error loading main configuration %v\n", err.Error())
	}
	log.Print(cfg)

	router := mux.NewRouter().StrictSlash(true)

	addOrderHandlers(router, cfg.DbType)
	addEventsHandler(router)
	addAPIHandler(router)

	if err := startService(cfg.Port, router); err != nil {
		log.Fatal("Unable to start server", err)
	}
}

func addOrderHandlers(router *mux.Router, dbType string) {

	/*repo, err := Create(dbType)
	if err != nil {
		log.Fatal("Unable to initiate repository", err)
	}

	//orderHandler := handler.NewOrderHandler(repo)*/

	// orders
	router.HandleFunc("/orders", r.InsertOrder).Methods(http.MethodPost)

	router.HandleFunc("/orders", r.GetOrders).Methods(http.MethodGet)
	router.HandleFunc("/namespace/{namespace}/orders", r.GetNamespaceOrders).Methods(http.MethodGet)

	router.HandleFunc("/orders", r.DeleteOrders).Methods(http.MethodDelete)
	router.HandleFunc("/namespace/{namespace}/orders", r.DeleteNamespaceOrders).Methods(http.MethodDelete)
}

func addEventsHandler(router *mux.Router) {
	router.HandleFunc("/events/order/created", events.HandleOrderCreatedEvent).Methods(http.MethodPost)

}

func addAPIHandler(router *mux.Router) {
	// API
	router.HandleFunc("/", handler.SwaggerAPIRedirectHandler).Methods(http.MethodGet)
	router.HandleFunc("/api.yaml", handler.SwaggerAPIHandler).Methods(http.MethodGet)
}

func startService(port string, router *mux.Router) error {
	log.Printf("Starting server on port %s ", port)

	c := cors.AllowAll()
	return http.ListenAndServe(":"+port, c.Handler(router))
}


// Create is used to create an OrderRepository based on the given dbtype.
// Currently the `MemoryDatabase` and `SQLServerDriverName` are supported.
/*func Create(dbtype string) (repository.OrderRepository, error) {

	var (
		dbCfg config.Config
		err error
	)
	if err = envconfig.Init(&dbCfg); err != nil {
		return nil, errors.Wrap(err, "Error loading db configuration %v.")
	}

	switch dbtype {
	case config.MemoryDatabase:
		return repository.NewOrderRepositoryMemory(), nil
	case config.SQLServerDriverName:
		mssql := mssqldb.Mssql{dbCfg}
		return mssql.NewOrderRepositoryDb()
	case config.PostgresDriverName:
		postgresDB := postgres.Postgres{dbCfg}
		return postgresDB.NewOrderRepositoryDb()
	default:
		return nil, errors.Errorf("Unsupported database type %s", dbtype)
	}
}*/