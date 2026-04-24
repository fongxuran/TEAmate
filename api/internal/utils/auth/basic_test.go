package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	users := map[string]string{"admin": "password"}
	middleware := Middleware(users)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	unauthorizedReq := httptest.NewRequest(http.MethodGet, "/", nil)
	unauthorizedRec := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedRec, unauthorizedReq)
	if unauthorizedRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", unauthorizedRec.Code)
	}

	authorizedReq := httptest.NewRequest(http.MethodGet, "/", nil)
	authorizedReq.SetBasicAuth("admin", "password")
	authorizedRec := httptest.NewRecorder()
	handler.ServeHTTP(authorizedRec, authorizedReq)
	if authorizedRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", authorizedRec.Code)
	}
}
