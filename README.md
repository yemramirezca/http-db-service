# HTTP DB Service

## Overview

This example demonstrates nd multi instance routing using Kyma capabilities, such as HTTP endpoints that expose and bind a service to a database. The service in this example exposes HTTP endpoints used to create and read basic order JSON entities, as described in the [service's API descriptor](docs/api/api.yaml). The service can run with a database Postgres instance.  The service in this example uses [Go](http://golang.org).

## Prerequisites

- A [Docker](https://docs.docker.com/install) installation.
- Kyma as the target deployment environment.
- An Postgres database for the service's database functionality.
- [Golang](https://golang.org/dl/) 

## Installation

Use these commands to build and run the service with Docker:

```
docker build -t http-db-service:latest .
docker run -it --rm -p 8017:8017 http-db-service:latest
```

To configure the connection to the database, set the environment variables for the values defined in the `config/service.go` file.

The `deployment` folder contains `.yaml` descriptors used for the deployment of the service to Kyma.

Run the following commands to deploy the published service to Kyma:

1. Export your Namespace as variable by replacing the `{namespace}` placeholder in the following command and running it:

    ```bash
    export KYMA_EXAMPLE_NS="{namespace}"
    ```
2. Deploy the service:
    ```bash
    kubectl create namespace $KYMA_EXAMPLE_NS
    kubectl apply -f deployment/http-db-service.yaml -n $KYMA_EXAMPLE_NS
    kubectl apply -f deployment/postgres-binding-usage.yaml -n $KYMA_EXAMPLE_NS
    ```

### Tests
Perform a request against ```$CLUSTER-DOMAIN/orders``` and set the header ```end-user``` 
See how the service uses a different database depending on the end-user


### Cleanup

Run the following command to completely remove the example and all its resources from the cluster:

```bash
kubectl delete all -l example=http-db-service -n $KYMA_EXAMPLE_NS
```
