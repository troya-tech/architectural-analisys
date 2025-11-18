package infrastructure

import (
    "sync"

    "modular-monolith/internal/todo/domain"
)

type memRepo struct {
    mu    sync.RWMutex
    next  int64
    items map[int64]domain.Todo
}

func NewMemRepo() *memRepo {
    return &memRepo{next: 1, items: make(map[int64]domain.Todo)}
}

func (r *memRepo) NextID() int64 {
    r.mu.Lock()
    defer r.mu.Unlock()
    id := r.next
    r.next++
    return id
}
func (r *memRepo) Save(t domain.Todo) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.items[t.ID] = t
    return nil
}
func (r *memRepo) FindByID(id int64) (domain.Todo, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    t, ok := r.items[id]
    return t, ok
}
func (r *memRepo) FindAll() []domain.Todo {
    r.mu.RLock()
    defer r.mu.RUnlock()
    out := make([]domain.Todo, 0, len(r.items))
    for _, t := range r.items {
        out = append(out, t)
    }
    return out
}
func (r *memRepo) Update(t domain.Todo) bool {
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
