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

func (u *usecase) ReadConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

func (u *usecase) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":   "OK",
		"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
	})
}
