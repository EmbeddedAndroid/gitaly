package config

import (
	log "github.com/Sirupsen/logrus"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

// ConfigurePrometheus uses the global configuration to configure prometheus
func ConfigurePrometheus() {
	if len(Config.Prometheus.GRPCLatencyBuckets) == 0 {
		return
	}

	log.WithField("latencies", Config.Prometheus.GRPCLatencyBuckets).Debug("grpc prometheus histograms enabled")

	grpc_prometheus.EnableHandlingTimeHistogram(func(histogramOpts *prometheus.HistogramOpts) {
		histogramOpts.Buckets = Config.Prometheus.GRPCLatencyBuckets
	})
}
