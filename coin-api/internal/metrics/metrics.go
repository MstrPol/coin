package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"coin.local/coin-api/internal/store"
)

var resolveDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "coin_resolve_duration_seconds",
		Help:    "Duration of manifest resolve requests",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"gp_name", "gp_version", "status"},
)

var scanDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "coin_scan_duration_seconds",
		Help:    "Duration of fleet scanner runs",
		Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600, 7200},
	},
)

var scanReposScanned = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "coin_repos_scanned",
		Help: "Repositories scanned in the last fleet scan",
	},
)

var scanReposTotal = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "coin_scan_repos_total",
		Help: "Total repositories discovered in the last fleet scan",
	},
)

var scanReposSkipped = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "coin_scan_repos_skipped",
		Help: "Repositories skipped in the last fleet scan",
	},
)

var scanReposFailed = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "coin_scan_repos_failed",
		Help: "Repositories failed in the last fleet scan",
	},
)

var scanLastSuccess = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "coin_scan_last_success_timestamp",
		Help: "Unix timestamp of the last successful fleet scan",
	},
)

func init() {
	prometheus.MustRegister(
		resolveDuration,
		scanDuration,
		scanReposScanned,
		scanReposTotal,
		scanReposSkipped,
		scanReposFailed,
		scanLastSuccess,
	)
}

func ObserveResolve(gpName, gpVersion, status string, d time.Duration) {
	resolveDuration.WithLabelValues(gpName, gpVersion, status).Observe(d.Seconds())
}

func ObserveScan(result store.ScanResult, err error) {
	if !result.FinishedAt.IsZero() {
		scanDuration.Observe(result.FinishedAt.Sub(result.StartedAt).Seconds())
	}
	scanReposTotal.Set(float64(result.ReposTotal))
	scanReposScanned.Set(float64(result.ReposScanned))
	scanReposSkipped.Set(float64(result.ReposSkipped))
	scanReposFailed.Set(float64(result.ReposFailed))
	if err == nil && result.ReposFailed == 0 {
		scanLastSuccess.Set(float64(result.FinishedAt.Unix()))
	}
}
