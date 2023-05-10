package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
	"testing"
)

// TODO: test metrics scenario
// why does that test needs app.yml?
func TestMetric(t *testing.T) {
	router := gin.Default()
	p := metrics.Startup(router)

	// Setup metric
	metricCounter := &ginPrometheus.Metric{
		ID:          MetricCounterTestID,   // optional string
		Name:        "test_metric",         // required string
		Description: "Counter test metric", // required string
		Type:        "counter",             // required string
	}
	metrics.RegisterCustomMetric(p, metricCounter)

	// inc
	m, ok := metrics.GetMetricByID(p, MetricCounterTestID)
	if ok {
		logger.Info("ПОЛУЧИЛОСЬ!", m)
		m.MetricCollector.(prometheus.Counter).Inc()
	}
}
