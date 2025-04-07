package handler

// Handler объединяет все обработчики
type Handler struct {
	Auth *AuthHandler
	Task *TaskHandler
}

// NewHandler создает новый экземпляр Handler
func NewHandler(auth *AuthHandler, task *TaskHandler) *Handler {
	return &Handler{
		Auth: auth,
		Task: task,
	}
}
