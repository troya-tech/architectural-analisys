package main

import (
    "log"
    "net/http"

    "modular-monolith/internal/todo/api"
    "modular-monolith/internal/todo/application"
    "modular-monolith/internal/todo/infrastructure"
)

func main() {
    repo := infrastructure.NewMemRepo()
    uc := application.NewUseCases(repo)

    mux := http.NewServeMux()
    api.RegisterRoutes(mux, uc)

    addr := ":8081"
    log.Printf("modular-monolith listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, mux))
}
