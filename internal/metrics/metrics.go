package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests partitioned by method, path and status code.",
		},
		[]string{"method", "path", "status_code"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency partitioned by method and path.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	HTTPResponseBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_bytes_total",
			Help: "Total bytes sent in HTTP responses partitioned by method and path.",
		},
		[]string{"method", "path"},
	)

	HTTPUniqueVisitorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_unique_visitors_total",
			Help: "Unique daily visitors identified by IP and user-agent.",
		},
	)

	ShopsCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "shophub_shops_created_total",
			Help: "Total number of shops successfully created via ShopHub.",
		},
	)
)

func Register() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPResponseBytesTotal,
		HTTPUniqueVisitorsTotal,
		ShopsCreatedTotal,
	)
}
