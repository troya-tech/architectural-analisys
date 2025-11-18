package application

import (
    "errors"
    "strings"

    "modular-monolith/internal/todo/domain"
)

type UseCases struct{ repo Repo }

func NewUseCases(r Repo) *UseCases { return &UseCases{repo: r} }

func (u *UseCases) Create(title string) (domain.Todo, error) {
    title = strings.TrimSpace(title)
    if title == "" {
        return domain.Todo{}, errors.New("title required")
    }
    id := u.repo.NextID()
    t := domain.NewTodo(id, title)
    return t, u.repo.Save(t)
}
func (u *UseCases) Get(id int64) (domain.Todo, bool)  { return u.repo.FindByID(id) }
func (u *UseCases) List() []domain.Todo               { return u.repo.FindAll() }
func (u *UseCases) Update(id int64, title string, completed bool) (domain.Todo, bool) {
    t, ok := u.repo.FindByID(id)
    if !ok {
        return domain.Todo{}, false
    }
    if s := strings.TrimSpace(title); s != "" {
        t.Title = s
    }
    t.Completed = completed
    ok = u.repo.Update(t)
    return t, ok
}
func (u *UseCases) Delete(id int64) bool { return u.repo.Delete(id) }
