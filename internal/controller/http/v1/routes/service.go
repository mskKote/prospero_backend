package routes

import (
	"github.com/gin-gonic/gin"
)

const (
	configURL = "/config"
)

type IServiceUsecase interface {
	ReadConfig(c *gin.Context)
}

func RegisterServiceRoutes(g *gin.RouterGroup, s IServiceUsecase) {
	g.GET(configURL, s.ReadConfig)
}
