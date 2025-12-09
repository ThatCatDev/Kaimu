package metrics

import (
	metricsLib "github.com/weeb-vip/go-metrics-lib"
	"github.com/weeb-vip/go-metrics-lib/clients/prometheus"
)

var metricsInstance metricsLib.MetricsImpl
var prometheusInstance *prometheus.PrometheusClient

func NewMetricsInstance() metricsLib.MetricsImpl {
	if metricsInstance == nil {
		prometheusInstance = NewPrometheusInstance()
		metricsInstance = metricsLib.NewMetrics(prometheusInstance, 1)
	}
	return metricsInstance
}

func NewPrometheusInstance() *prometheus.PrometheusClient {
	if prometheusInstance == nil {
		prometheusInstance = prometheus.NewPrometheusClient()
	}
	return prometheusInstance
}
