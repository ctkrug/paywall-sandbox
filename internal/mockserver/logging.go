package mockserver

import (
	"log"
	"net/http"
	"time"
)

// LogRequests wraps next with structured request logging: method, path,
// resulting status, and latency. This is what --verbose surfaces so a user
// can see exactly which requests were challenged versus served.
func LogRequests(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, req)
		logger.Printf("%s %s -> %d (%s)", req.Method, req.URL.Path, rec.status, time.Since(start))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
