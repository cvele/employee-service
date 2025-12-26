package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsProvider struct {
	Seconds  *prometheus.HistogramVec
	Requests *prometheus.CounterVec
}

func NewMetricsProvider(namespace, subsystem string) *MetricsProvider {
	seconds := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_duration_seconds",
		Help:      "Request latencies in seconds.",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
	}, []string{"kind", "operation", "code", "reason"})

	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_total",
		Help:      "Total number of requests.",
	}, []string{"kind", "operation", "code", "reason"})

	prometheus.MustRegister(seconds, requests)

	return &MetricsProvider{
		Seconds:  seconds,
		Requests: requests,
	}
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

