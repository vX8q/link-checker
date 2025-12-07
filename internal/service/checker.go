package service

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// checkLinks выполняет проверку ссылок из задачи `t`.
// Для каждой ссылки формируется HEAD-запрос (если в начале нет схемы,
// добавляется "http://"). По коду ответа определяем статус: "available" для
// кодов < 400, иначе "not available". Результаты сохраняются в хранилище
// под исходной ссылкой (без автоматически добавленной схемы).
func (s *TaskService) checkLinks(t *Task) {
	// настроим транспорт с более агрессивными таймаутами
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 0,
		}).Dial,
		IdleConnTimeout: 1 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   7 * time.Second,
	}
	defer client.CloseIdleConnections()

	for _, link := range t.Links {
		url := link
		if !strings.HasPrefix(url, "http") {
			url = "http://" + url
		}

		status := "not available"

		// используем контекст с timeout
		ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
		req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
		if err != nil {
			log.Printf("failed to create request for %s: %v", url, err)
			cancel()
			s.st.SaveResult(t.ID, link, status)
			continue
		}

		resp, err := client.Do(req)
		cancel()

		if err != nil {
			log.Printf("request failed for %s: %v", url, err)
			s.st.SaveResult(t.ID, link, status)
			continue
		}

		if resp.StatusCode < 400 {
			status = "available"
		}
		resp.Body.Close()

		// сохраняем результат под исходной ссылкой (link, а не url)
		s.st.SaveResult(t.ID, link, status)
	}
}
