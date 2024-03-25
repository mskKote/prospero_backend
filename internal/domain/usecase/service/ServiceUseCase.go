package service

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/pkg/config"
	"net/http"
	"time"
)

var cfg = config.GetConfig()

// usecase - зависимые сервисы
type usecase struct {
}

func New() IServiceUseCase {
	return &usecase{}
}

// ReadConfig godoc
//
//	@Summary		Get service config
//	@Description	Full service config (env + app.yml)
//	@Tags			service
//	@Produce		json
//	@Accept			json
//	@Success		200	{object}	config.Config
//	@Router			/service/config [get]
func (u *usecase) ReadConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

// HealthCheck godoc
//
// HealthCheck godoc
//
//	@Summary		Perform health check
//	@Description	Check if the service is healthy
//	@Tags			service
//	@Produce		json
//	@Success		200
//	@Router			/service/healthcheck [get]
func (u *usecase) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":   "OK",
		"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
	})
}
