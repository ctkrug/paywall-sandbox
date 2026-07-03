package mockserver

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogRequestsRecordsStatusAndPath(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := LogRequests(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
	}))

	req := httptest.NewRequest("GET", "/paid", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	got := buf.String()
	if !strings.Contains(got, "GET /paid -> 402") {
		t.Fatalf("log output = %q, want it to contain %q", got, "GET /paid -> 402")
	}
}

func TestLogRequestsDefaultsToOKWhenUnset(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	handler := LogRequests(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/free", nil))

	if !strings.Contains(buf.String(), "GET /free -> 200") {
		t.Fatalf("log output = %q, want it to contain %q", buf.String(), "GET /free -> 200")
	}
}
