package config

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const (
	metricNamespace   = "nodechecker_controller"
	metricsServerPort = ":2112"
)

type ServerMetrics struct {
	metrics map[string]prometheus.Collector
}

func StartMetricServer() *ServerMetrics {

	log := logrus.WithFields(logrus.Fields{"Node": os.Getenv("NODE_NAME")})

	metrics := &ServerMetrics{}
	metrics.CreateMetrics()

	go func() {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())
		log.Infof("Starting metric server at address [%s]", metricsServerPort)
		if err := http.ListenAndServe(metricsServerPort, metricsMux); err != nil {
			log.Errorf("Error to start the metric server: %v", err.Error())
		}
	}()

	return metrics
}

func (m *ServerMetrics) CreateMetrics() {
	NodeCheckerConnectionError := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "nodechecker_connection_error_total",
			Help:      "The total number of connection error",
		}, []string{"node", "rule", "type", "destination", "schedule"})

	NodeCheckerFeatureError := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Name:      "nodechecker_feature_error_total",
			Help:      "The total number of feature error",
		}, []string{"node", "feature", "schedule"})

	// Insert metrics
	m.metrics = map[string]prometheus.Collector{
		"nodeCheckerConnectionError": NodeCheckerConnectionError,
		"nodeCheckerFeatureError":    NodeCheckerFeatureError,
	}

	// Register all metrics
	for _, i := range m.metrics {
		prometheus.MustRegister(i)
	}

}

func (m *ServerMetrics) IncConnectionCheckError(node string, rule string, typeCheck string, destination string, schedule string) {
	if c, ok := m.metrics["nodeCheckerConnectionError"].(*prometheus.GaugeVec); ok {
		c.WithLabelValues(node, rule, typeCheck, destination, schedule).Inc()
	}
}

func (m *ServerMetrics) SetConnectionCheckError(node string, rule string, typeCheck string, destination string, schedule string, value float64) {
	if c, ok := m.metrics["nodeCheckerConnectionError"].(*prometheus.GaugeVec); ok {
		c.WithLabelValues(node, rule, typeCheck, destination, schedule).Set(value)
	}
}

func (m *ServerMetrics) IncFeatureCheckError(node string, feature string, schedule string) {
	if c, ok := m.metrics["nodeCheckerFeatureError"].(*prometheus.GaugeVec); ok {
		c.WithLabelValues(node, feature, schedule).Inc()
	}
}

func (m *ServerMetrics) SetFeatureCheckError(node string, feature string, schedule string, value float64) {
	if c, ok := m.metrics["nodeCheckerFeatureError"].(*prometheus.GaugeVec); ok {
		c.WithLabelValues(node, feature, schedule).Set(value)
	}
}
