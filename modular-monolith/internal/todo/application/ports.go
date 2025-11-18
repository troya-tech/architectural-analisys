package application

import "modular-monolith/internal/todo/domain"

type Repo interface {
    NextID() int64
    Save(t domain.Todo) error
    FindByID(id int64) (domain.Todo, bool)
    FindAll() []domain.Todo
    Update(t domain.Todo) bool
    Delete(id int64) bool
}
