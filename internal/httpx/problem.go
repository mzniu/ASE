package httpx

import (
	"encoding/json"
	"log"
	"net/http"
)

// ProblemDetail is a minimal RFC 7807-style body.
type ProblemDetail struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Status int    `json:"status"`
}

// WriteProblem writes application/problem+json.
func WriteProblem(w http.ResponseWriter, status int, title, detail string) {
	body, err := json.Marshal(ProblemDetail{
		Type:   "about:blank",
		Title:  title,
		Detail: detail,
		Status: status,
	})
	if err != nil {
		log.Printf("problem json: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		log.Printf("write response: %v", err)
	}
}
