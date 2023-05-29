package metrics

import (
	"github.com/mskKote/prospero_backend/pkg/metrics"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
)

const (
	MetricCounterTestName = "test_metric"
	MetricSummaryTestName = "test_metric_2"
	MetricRssObtainName   = "metric_rss_harvest"
)

func RegisterMetrics(p *ginPrometheus.Prometheus) {
	metricCounter := &ginPrometheus.Metric{
		Name:        MetricCounterTestName, // required string
		Description: "Counter test metric", // required string
		Type:        "counter",             // required string
	}
	metrics.RegisterCustomMetric(p, metricCounter)

	metricSummary := &ginPrometheus.Metric{
		Name:        MetricSummaryTestName, // Metric Name
		Description: "Summary test metric", // Help Description
		Type:        "summary",             // type associated with prometheus collector
	}
	metrics.RegisterCustomMetric(p, metricSummary)

	// Время прохода по RSS источникам
	metricRssHarvestSummary := &ginPrometheus.Metric{
		Name:        MetricRssObtainName,               // Metric Name
		Description: "Время прохода по RSS источникам", // Help Description
		Type:        "summary",                         // type associated with prometheus collector
	}
	metrics.RegisterCustomMetric(p, metricRssHarvestSummary)
}
