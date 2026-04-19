package handler

import "net/http"

// Health is a liveness probe for orchestrators and load balancers (no auth, not rate-limited).
func Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
