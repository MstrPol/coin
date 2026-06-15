package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var resolveDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "coin_resolve_duration_seconds",
		Help:    "Duration of manifest resolve requests",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"gp_name", "gp_version", "status"},
)

func init() {
	prometheus.MustRegister(resolveDuration)
}

func ObserveResolve(gpName, gpVersion, status string, d time.Duration) {
	resolveDuration.WithLabelValues(gpName, gpVersion, status).Observe(d.Seconds())
}
