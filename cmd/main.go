package main

import (
	"github.com/gin-gonic/gin"
	internalMetrics "github.com/mskKote/prospero_backend/internal/adapters/metrics"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/routes"
	"github.com/mskKote/prospero_backend/internal/domain/usecase/search"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	pkgMetrics "github.com/mskKote/prospero_backend/pkg/metrics"
)

var (
	logger = logging.GetLogger()
	cfg    = config.GetConfig()
)

func main() {
	logger.Info("logger is OK")
	startup(cfg)
}

func startup(cfg *config.Config) {
	router := gin.New() // empty engine
	if cfg.IsDebug == false {
		gin.SetMode(gin.ReleaseMode)
	}

	// --------------------------------------- MIDDLEWARE
	// Recovery
	router.Use(gin.Recovery())

	// Logging
	if cfg.Logger.UseDefaultGin {
		logger.Info("Используем DefaultGin")
		router.Use(gin.Logger())
	}
	if cfg.Logger.ToGraylog {
		logger.Info("Используем Graylog")
		router.Use(logging.GraylogMiddlewareLogger())
	}
	// Metrics
	p := pkgMetrics.Startup(router)
	internalMetrics.RegisterMetrics(p)

	// --------------------------------------- ROUTES
	apiV1 := router.Group("/api/v1")
	routes.
		NewSearchRoute(&search.Usecase{}).
		Register(apiV1)

	// --------------------------------------- IGNITE
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Fatalln("ошибка, завершаем программу", err)
	}
}
