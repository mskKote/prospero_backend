package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"github.com/prometheus/client_golang/prometheus"
	ginPrometheus "github.com/zsais/go-gin-prometheus"
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
		logger.WithError(err).Errorf("%s could not be registered in Prometheus", m.Name)
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
