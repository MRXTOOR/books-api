package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggerMiddleware(t *testing.T) {
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
