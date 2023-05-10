package metrics

import (
	"github.com/mskKote/prospero_backend/pkg/logging"
	"github.com/mskKote/prospero_backend/pkg/metrics"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
)

const (
	MetricCounterTestID = "1234"
	MetricSummaryTestID = "1235"
)

var logger = logging.GetLogger()

func RegisterMetrics(p *ginPrometheus.Prometheus) {
	metricCounter := &ginPrometheus.Metric{
		ID:          MetricCounterTestID,   // optional string
		Name:        "test_metric",         // required string
		Description: "Counter test metric", // required string
		Type:        "counter",             // required string
	}
	metrics.RegisterCustomMetric(p, metricCounter)

	metricSummary := &ginPrometheus.Metric{
		ID:          MetricSummaryTestID,   // Identifier
		Name:        "test_metric_2",       // Metric Name
		Description: "Summary test metric", // Help Description
		Type:        "summary",             // type associated with prometheus collector
	}
	metrics.RegisterCustomMetric(p, metricSummary)
}
