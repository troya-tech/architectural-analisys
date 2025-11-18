package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "sync"
)

type Todo struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}

type Repo interface {
    NextID() int64
    Save(Todo)
    Get(id int64) (Todo, bool)
    List() []Todo
    Update(Todo) bool
    Delete(id int64) bool
}

type memRepo struct {
    mu    sync.RWMutex
    next  int64
    items map[int64]Todo
}

func newMemRepo() *memRepo { return &memRepo{next: 1, items: make(map[int64]Todo)} }
func (r *memRepo) NextID() int64 {
    r.mu.Lock()
    defer r.mu.Unlock()
    id := r.next
    r.next++
    return id
}
func (r *memRepo) Save(t Todo) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.items[t.ID] = t
}
func (r *memRepo) Get(id int64) (Todo, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    t, ok := r.items[id]
    return t, ok
}
func (r *memRepo) List() []Todo {
    r.mu.RLock()
    defer r.mu.RUnlock()
    out := make([]Todo, 0, len(r.items))
    for _, t := range r.items {
        out = append(out, t)
    }
    return out
}
func (r *memRepo) Update(t Todo) bool {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.items[t.ID]; !ok {
        return false
    }
    r.items[t.ID] = t
    return true
}
func (r *memRepo) Delete(id int64) bool {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.items[id]; !ok {
        return false
    }
    delete(r.items, id)
    return true
}

type UseCases struct{ repo Repo }

func newUseCases(r Repo) *UseCases { return &UseCases{repo: r} }

func (u *UseCases) Create(title string) (Todo, bool) {
    title = strings.TrimSpace(title)
    if title == "" {
        return Todo{}, false
    }
    id := u.repo.NextID()
    t := Todo{ID: id, Title: title}
    u.repo.Save(t)
    return t, true
}
func (u *UseCases) Get(id int64) (Todo, bool)                       { return u.repo.Get(id) }
func (u *UseCases) List() []Todo                                    { return u.repo.List() }
func (u *UseCases) Update(id int64, title string, completed bool) (Todo, bool) {
    t, ok := u.repo.Get(id)
    if !ok {
        return Todo{}, false
    }
    if s := strings.TrimSpace(title); s != "" {
        t.Title = s
    }
    t.Completed = completed
    return t, u.repo.Update(t)
}
func (u *UseCases) Delete(id int64) bool { return u.repo.Delete(id) }

func main() {
    repo := newMemRepo()
    uc := newUseCases(repo)

    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
    mux.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            writeJSON(w, http.StatusOK, uc.List())
        case http.MethodPost:
            var in struct{ Title string }
            if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Title) == "" {
                http.Error(w, "bad request", http.StatusBadRequest)
                return
            }
            if t, ok := uc.Create(in.Title); ok {
                writeJSON(w, http.StatusCreated, t)
            } else {
                http.Error(w, "validation failed", http.StatusBadRequest)
            }
        default:
            w.WriteHeader(http.StatusMethodNotAllowed)
        }
    })
    mux.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
        id, ok := parseID(r.URL.Path)
        if !ok {
            http.NotFound(w, r)
            return
        }
        switch r.Method {
        case http.MethodGet:
            t, found := uc.Get(id)
            if !found {
                http.NotFound(w, r)
                return
            }
            writeJSON(w, http.StatusOK, t)
        case http.MethodPut:
            var in struct {
                Title     string `json:"title"`
                Completed bool   `json:"completed"`
            }
            if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
                http.Error(w, "bad request", http.StatusBadRequest)
                return
            }
            t, updated := uc.Update(id, in.Title, in.Completed)
            if !updated {
                http.NotFound(w, r)
                return
            }
            writeJSON(w, http.StatusOK, t)
        case http.MethodDelete:
            if ok := uc.Delete(id); !ok {
                http.NotFound(w, r)
                return
            }
            w.WriteHeader(http.StatusNoContent)
        default:
            w.WriteHeader(http.StatusMethodNotAllowed)
        }
    })

    addr := ":" + env("PORT", "8082")
    log.Printf("todo-microservice listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, mux))
}

func env(k, def string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return def
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
