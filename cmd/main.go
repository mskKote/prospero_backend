package main

import (
	"github.com/gin-gonic/gin"
	internalMetrics "github.com/mskKote/prospero_backend/internal/adapters/metrics"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/routes"
	"github.com/mskKote/prospero_backend/internal/domain/usecase/search"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	pkgMetrics "github.com/mskKote/prospero_backend/pkg/metrics"
	"go.uber.org/zap"
)

var (
	cfg    = config.GetConfig()
	logger = logging.GetLogger()
)

func main() {
	startup(cfg)
}

func startup(cfg *config.Config) {

	r := gin.New() // empty engine
	if cfg.IsDebug == false {
		gin.SetMode(gin.ReleaseMode)
	}

	// --------------------------------------- MIDDLEWARE
	// Recovery
	r.Use(gin.Recovery())

	// Logging
	if cfg.Logger.UseDefaultGin {
		logger.Info("Используем DefaultGin")
		r.Use(gin.Logger())
	}
	//if cfg.Logger.ToGraylog { // logrus & graylog
	//	logger.Info("Используем Graylog")
	//	r.Use(logging.GraylogMiddlewareLogger())
	//}
	if cfg.Logger.UseZap {
		logger.Info("Используем Zap")
		logging.ZapMiddlewareLogger(r)
	}
	// Metrics
	p := pkgMetrics.Startup(r)
	internalMetrics.RegisterMetrics(p)

	// --------------------------------------- ROUTES
	apiV1 := r.Group("/api/v1")
	routes.
		NewSearchRoute(&search.Usecase{}).
		Register(apiV1)

	// --------------------------------------- IGNITE
	if err := r.Run(":" + cfg.Port); err != nil {
		logger.Fatal("ошибка, завершаем программу", zap.Error(err))
	}
}
