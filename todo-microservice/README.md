# Todo Microservice

Single service with health endpoint and Todo CRUD.

## Run (local)

```
cd c:\DEV\Playground\craftmanship\todo-microservice
go run .
```

## Docker build & run

```
cd c:\DEV\Playground\craftmanship\todo-microservice
docker build -t todo-microservice:local .
docker run --rm -p 8082:8080 -e PORT=8080 todo-microservice:local
```

Base URL: `http://localhost:8082` when using `go run .` (PORT=8082 default).