package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/routes"
	"github.com/mskKote/prospero_backend/internal/domain/usecase/search"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
)

var (
	logger = logging.GetLogger()
)

func main() {
	router := gin.Default()
	cfg := config.GetConfig()
	logger.Info("Прочитали config, запускаемся")
	startup(router, cfg)
}

func startup(router *gin.Engine, cfg *config.Config) {

	if *cfg.IsDebug == false {
		gin.SetMode(gin.ReleaseMode)
	}
	apiV1 := router.Group("/api/v1")
	routes.
		NewSearchRoute(&search.Usecase{}).
		Register(apiV1)

	if err := router.Run(":" + cfg.Listen.Port); err != nil {
		logger.Fatalln("ошибка, завершаем программу", err)
	}
}
