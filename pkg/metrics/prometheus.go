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
	logger = logging.GetLogger()
	cfg    = config.GetConfig()
)

func GetMetricByID(p *ginPrometheus.Prometheus, id string) (*ginPrometheus.Metric, bool) {
	for i := range p.MetricsList {
		if p.MetricsList[i].ID == id {
			a := p.MetricsList[i]
			return a, true
		}
	}
	return nil, false
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
	// gin middleware
	p.Use(router)
	return p
}
