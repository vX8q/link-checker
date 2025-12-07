package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"link/internal/service"

	"github.com/julienschmidt/httprouter"
)

// Handler предоставляет HTTP-обработчики для работы с задачами проверки ссылок.
type Handler struct {
	service *service.TaskService
}

// New создает и возвращает новый экземпляр Handler, инициализированный предоставленным TaskService

func New(service *service.TaskService) *Handler {
	return &Handler{service: service}
}

// LinksRequest определяет структуру для тела JSON-запроса, содержащего список ссылок

type LinksRequest struct {
	Links []string `json:"links"`
}

// Links обрабатывает POST-запрос для отправки ссылок.
// Создаёт задачу, ждёт результатов проверки (до 10 секунд)
// и возвращает JSON-объект с ID задачи и статусами ссылок.
func (h *Handler) Links(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req LinksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		log.Printf("failed to decode request body: %v", err)
		return
	}
	defer r.Body.Close()

	if len(req.Links) == 0 {
		http.Error(w, "no links provided", http.StatusBadRequest)
		return
	}

	id, _ := h.service.Create(req.Links)

	// ждём результатов проверки (максимум 10 секунд)
	statuses := h.service.WaitAndGetResults(id, 10*time.Second)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"links":     statuses,
		"links_num": id,
	})
}
