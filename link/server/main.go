package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"link/internal/app"
)

// main запускает HTTP-сервер, настраивает обработчики приложения и
// обеспечивает корректное завершение работы при получении сигналов ОС.
func main() {
	a := app.New()

	srv := &http.Server{
		Addr:    ":8000",
		Handler: a.Router,
	}

	go func() {
		log.Println("server started on :8000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Ожидаем сигнала завершения (Ctrl+C или kill)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	sig := <-stop
	log.Printf("received signal %v, shutting down...", sig)

	// Даем время на graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	a.Shutdown()
	log.Println("server stopped")
}
