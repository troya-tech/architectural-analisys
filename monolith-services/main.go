package main

import (
    "encoding/json"
    "errors"
    "log"
    "net/http"
    "strconv"
    "strings"
    "sync"
)

type Todo struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}

type TodoRepo interface {
    NextID() int64
    Save(t Todo) error
    FindByID(id int64) (Todo, error)
    FindAll() ([]Todo, error)
    Update(t Todo) error
    Delete(id int64) error
}

type memRepo struct {
    mu    sync.RWMutex
    next  int64
    items map[int64]Todo
}

func newMemRepo() *memRepo {
    return &memRepo{next: 1, items: make(map[int64]Todo)}
}
func (r *memRepo) NextID() int64 {
    r.mu.Lock()
    defer r.mu.Unlock()
    id := r.next
    r.next++
    return id
}
func (r *memRepo) Save(t Todo) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.items[t.ID] = t
    return nil
}
func (r *memRepo) FindByID(id int64) (Todo, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    t, ok := r.items[id]
    if !ok {
        return Todo{}, errors.New("not found")
    }
    return t, nil
}
func (r *memRepo) FindAll() ([]Todo, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    out := make([]Todo, 0, len(r.items))
    for _, t := range r.items {
        out = append(out, t)
    }
    return out, nil
}
func (r *memRepo) Update(t Todo) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.items[t.ID]; !ok {
        return errors.New("not found")
    }
    r.items[t.ID] = t
    return nil
}
func (r *memRepo) Delete(id int64) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.items[id]; !ok {
        return errors.New("not found")
    }
    delete(r.items, id)
    return nil
}

type TodoService struct{ repo TodoRepo }

func NewTodoService(r TodoRepo) *TodoService { return &TodoService{repo: r} }

func (s *TodoService) Create(title string) (Todo, error) {
    title = strings.TrimSpace(title)
    if title == "" {
        return Todo{}, errors.New("title required")
    }
    id := s.repo.NextID()
    t := Todo{ID: id, Title: title, Completed: false}
    return t, s.repo.Save(t)
}
func (s *TodoService) Get(id int64) (Todo, error)        { return s.repo.FindByID(id) }
func (s *TodoService) List() ([]Todo, error)             { return s.repo.FindAll() }
func (s *TodoService) Update(id int64, title string, completed bool) (Todo, error) {
    t, err := s.repo.FindByID(id)
    if err != nil {
        return Todo{}, err
    }
    if sTitle := strings.TrimSpace(title); sTitle != "" {
        t.Title = sTitle
    }
    t.Completed = completed
    return t, s.repo.Update(t)
}
func (s *TodoService) Delete(id int64) error { return s.repo.Delete(id) }

func main() {
    repo := newMemRepo()
    svc := NewTodoService(repo)

    mux := http.NewServeMux()

    mux.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            todos, _ := svc.List()
            writeJSON(w, http.StatusOK, todos)
        case http.MethodPost:
            var in struct {
                Title string `json:"title"`
            }
            if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Title) == "" {
                http.Error(w, "invalid body", http.StatusBadRequest)
                return
            }
            t, err := svc.Create(in.Title)
            if err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }
            writeJSON(w, http.StatusCreated, t)
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
            t, err := svc.Get(id)
            if err != nil {
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
                http.Error(w, "invalid body", http.StatusBadRequest)
                return
            }
            t, err := svc.Update(id, in.Title, in.Completed)
            if err != nil {
                http.NotFound(w, r)
                return
            }
            writeJSON(w, http.StatusOK, t)
        case http.MethodDelete:
            if err := svc.Delete(id); err != nil {
                http.NotFound(w, r)
                return
            }
            w.WriteHeader(http.StatusNoContent)
        default:
            w.WriteHeader(http.StatusMethodNotAllowed)
        }
    })

    addr := ":8080"
    log.Printf("monolith-services listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, mux))
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
