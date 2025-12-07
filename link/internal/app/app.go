package app

import (
	"link/internal/handler"
	"link/internal/service"
	"link/internal/storage"

	"github.com/julienschmidt/httprouter"
)

// App инкапсулирует основной маршрутизатор и сервис задач приложения.
type App struct {
	Router *httprouter.Router
	srv    *service.TaskService
}

// New создает и возвращает новый экземпляр App, настраивает хранилище,
// сервис задач, обработчик и регистрирует API-эндпоинты в HTTP-роутере.
func New() *App {
	st := storage.New()
	srv := service.NewTaskService(st)

	h := handler.New(srv)

	r := httprouter.New()
	r.POST("/api/links", h.Links)
	r.POST("/api/report", h.Report)

	return &App{
		Router: r,
		srv:    srv,
	}
}

// Shutdown корректно завершает работу приложения, остановив все фоновые воркеры.
func (a *App) Shutdown() {
	a.srv.Shutdown()
}
