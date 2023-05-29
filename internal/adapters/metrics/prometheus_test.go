package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/pkg/metrics"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
	"testing"
)

// EXAMPLE test metrics scenario
func TestMetric(t *testing.T) {
	router := gin.Default()
	p := metrics.Startup(router)

	// Setup metric
	metricCounter := &ginPrometheus.Metric{
		Name:        MetricCounterTestName, // required string
		Description: "Counter test metric", // required string
		Type:        "counter",             // required string
	}
	metrics.RegisterCustomMetric(p, metricCounter)

	// inc
	metrics.IncMetric(MetricCounterTestName)
}
