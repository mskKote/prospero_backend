package routes

import (
	"github.com/gin-gonic/gin"
)

const (
	configURL      = "/config"
	healthcheckURL = "/healthcheck"
)

type IServiceUseCase interface {
	ReadConfig(c *gin.Context)
	HealthCheck(c *gin.Context)
}

func RegisterServiceRoutes(g *gin.RouterGroup, s IServiceUseCase) {
	g.GET(configURL, s.ReadConfig)
	g.Any(healthcheckURL, s.HealthCheck)
}
