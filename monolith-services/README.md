# Monolith - Services Style

Minimal Todo CRUD in a single process using a simple service.

## Run

```
cd c:\DEV\Playground\craftmanship\monolith-services
go run .
```

- POST `http://localhost:8080/todos` body: `{ "title": "Test" }`
- GET `http://localhost:8080/todos`
- GET `http://localhost:8080/todos/1`
- PUT `http://localhost:8080/todos/1` body: `{ "title": "Updated", "completed": true }`
- DELETE `http://localhost:8080/todos/1`
