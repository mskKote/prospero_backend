package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/internal/config"
	"github.com/mskKote/prospero_backend/internal/searcher"
	"github.com/mskKote/prospero_backend/pkg/logging"
)

var (
	logger = logging.GetLogger()
)

func main() {
	router := gin.Default()
	cfg := config.GetConfig()
	logger.Info("Запустили сервер")
	startup(router, cfg)
}

func startup(router *gin.Engine, cfg *config.Config) {
	apiV1 := router.Group("/api/v1")
	searcher.New(logger).Register(apiV1)

	if err := router.Run(":" + cfg.Listen.Port); err != nil {
		logger.Fatalln("ошибка, завершаем программу", err)
	}
}
