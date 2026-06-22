package httpapi

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shophub-platform/shophub/internal/metrics"
)

var (
	visitorMu    sync.RWMutex
	seenVisitors = make(map[string]struct{})
)

// MetricsMiddleware records Prometheus HTTP metrics for each request.
// Must wrap the ServeMux so that r.Pattern (Go 1.22+) is populated by the
// time we read it after next.ServeHTTP returns.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		iw := &metricsWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(iw, r)

		// Go 1.22+ ServeMux sets r.Pattern to the registered route pattern
		// (e.g. "GET /api/v1/shops/{id}"). Strip the method prefix so the
		// label contains only the path part.
		pattern := stripMethod(r.Pattern)
		if pattern == "" {
			pattern = r.URL.Path
		}

		dur := time.Since(start).Seconds()
		statusStr := strconv.Itoa(iw.status)

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, pattern, statusStr).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, pattern).Observe(dur)
		metrics.HTTPResponseBytesTotal.WithLabelValues(r.Method, pattern).Add(float64(iw.bytesWritten))

		trackShophubVisitor(r)
	})
}

// stripMethod removes the "METHOD " prefix that Go 1.22+ includes in r.Pattern.
func stripMethod(pattern string) string {
	if idx := strings.Index(pattern, " "); idx >= 0 {
		return pattern[idx+1:]
	}
	return pattern
}

func trackShophubVisitor(r *http.Request) {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	ua := r.Header.Get("User-Agent")
	day := time.Now().UTC().Format("2006-01-02")
	raw := fmt.Sprintf("%s|%s|%s", ip, ua, day)
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))

	visitorMu.RLock()
	_, seen := seenVisitors[key]
	visitorMu.RUnlock()

	if !seen {
		visitorMu.Lock()
		if _, seen = seenVisitors[key]; !seen {
			seenVisitors[key] = struct{}{}
			metrics.HTTPUniqueVisitorsTotal.Inc()
		}
		visitorMu.Unlock()
	}
}

type metricsWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
}

func (mw *metricsWriter) WriteHeader(status int) {
	mw.status = status
	mw.ResponseWriter.WriteHeader(status)
}

func (mw *metricsWriter) Write(b []byte) (int, error) {
	n, err := mw.ResponseWriter.Write(b)
	mw.bytesWritten += n
	return n, err
}
