package domain

type Todo struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}

func NewTodo(id int64, title string) Todo {
    return Todo{ID: id, Title: title, Completed: false}
}
