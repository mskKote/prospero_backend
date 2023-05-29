package metrics

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"github.com/prometheus/client_golang/prometheus"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
	"go.uber.org/zap"
)

var (
	logger      = logging.GetLogger()
	cfg         = config.GetConfig()
	prometheus_ *ginPrometheus.Prometheus
)

func GetMetricByName(name string) (*ginPrometheus.Metric, bool) {
	for i := range prometheus_.MetricsList {
		if prometheus_.MetricsList[i].Name == name {
			a := prometheus_.MetricsList[i]
			return a, true
		}
	}
	return nil, false
}

func IncMetric(name string) {
	if m, ok := GetMetricByName(name); ok {
		m.MetricCollector.(prometheus.Counter).Inc()
	} else {
		logger.Error(fmt.Sprintf("Нет метрики [%s]", name))
	}
}

//func SetGaugeMetric(name string, value float64) {
//	if m, ok := GetMetricByName(name); ok {
//		m.MetricCollector.(prometheus.Gauge).Set(value)
//	} else {
//		logger.Error(fmt.Sprintf("Нет метрики [%s]", name))
//	}
//}

func ObserveSummaryMetric(name string, value float64) {
	if m, ok := GetMetricByName(name); ok {
		m.MetricCollector.(prometheus.Summary).Observe(value)
	} else {
		logger.Error(fmt.Sprintf("Нет метрики [%s]", name))
	}
}

func RegisterCustomMetric(
	p *ginPrometheus.Prometheus,
	m *ginPrometheus.Metric) {
	metricCollector := ginPrometheus.NewMetric(m, cfg.Service)

	if err := prometheus.Register(metricCollector); err != nil {
		logger.Error(fmt.Sprintf("[METRICS] could not be registered in Prometheus %s", m.Name), zap.Error(err))
	}

	m.MetricCollector = metricCollector
	p.MetricsList = append(p.MetricsList, m)
}

func Startup(router *gin.Engine) *ginPrometheus.Prometheus {
	p := ginPrometheus.NewPrometheus(cfg.Service)
	prometheus_ = p
	// gin middleware
	p.Use(router)
	return p
}
