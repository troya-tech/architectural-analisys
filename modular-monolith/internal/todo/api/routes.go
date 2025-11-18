package api

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"

    "modular-monolith/internal/todo/application"
)

type routes struct{ uc *application.UseCases }

func RegisterRoutes(mux *http.ServeMux, uc *application.UseCases) {
    r := &routes{uc: uc}
    mux.HandleFunc("/todos", r.collection)
    mux.HandleFunc("/todos/", r.item)
}

func (r *routes) collection(w http.ResponseWriter, req *http.Request) {
    switch req.Method {
    case http.MethodGet:
        writeJSON(w, http.StatusOK, r.uc.List())
    case http.MethodPost:
        var in struct{ Title string }
        if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
            http.Error(w, "bad request", http.StatusBadRequest)
            return
        }
        t, err := r.uc.Create(in.Title)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        writeJSON(w, http.StatusCreated, t)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func (r *routes) item(w http.ResponseWriter, req *http.Request) {
    id, ok := parseID(req.URL.Path)
    if !ok {
        http.NotFound(w, req)
        return
    }
    switch req.Method {
    case http.MethodGet:
        t, found := r.uc.Get(id)
        if !found {
            http.NotFound(w, req)
            return
        }
        writeJSON(w, http.StatusOK, t)
    case http.MethodPut:
        var in struct {
            Title     string `json:"title"`
            Completed bool   `json:"completed"`
        }
        if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
            http.Error(w, "bad request", http.StatusBadRequest)
            return
        }
        t, updated := r.uc.Update(id, in.Title, in.Completed)
        if !updated {
            http.NotFound(w, req)
            return
        }
        writeJSON(w, http.StatusOK, t)
    case http.MethodDelete:
        if ok := r.uc.Delete(id); !ok {
            http.NotFound(w, req)
            return
        }
        w.WriteHeader(http.StatusNoContent)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func parseID(path string) (int64, bool) {
    parts := strings.Split(strings.Trim(path, "/"), "/")
    if len(parts) != 2 || parts[0] != "todos" {
        return 0, false
    }
    id, err := strconv.ParseInt(parts[1], 10, 64)
    return id, err == nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
}
