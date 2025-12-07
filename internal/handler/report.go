package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"link/pdf"

	"github.com/julienschmidt/httprouter"
)

type ReportRequest struct {
	LinksList []int `json:"links_list"`
}

// Report генерирует PDF-отчёт по переданным ID задач и возвращает его как
// вложение в ответе (Content-Disposition: attachment).
func (h *Handler) Report(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		log.Printf("failed to decode request body: %v", err)
		return
	}
	defer r.Body.Close()

	data := h.service.GetForReport(req.LinksList)
	pdfBytes := pdf.Generate(data)

	// задаём заголовки ответа для типа содержимого PDF и скачиваемого вложения
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
	if _, err := w.Write(pdfBytes); err != nil {
		log.Printf("failed to write PDF response: %v", err)
	}
}
