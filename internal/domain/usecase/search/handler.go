package search

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"net/http"
)

var (
	logger = logging.GetLogger()
)

// TODO: сервисы к ElasticSearch
type Service interface {
	GrandFilter(g *gin.Context)
}

type Usecase struct {
	Service
}

func (h *Usecase) GrandFilter(g *gin.Context) {
	logger.Info("Вызвали GrandFilter")
	g.JSON(http.StatusOK, gin.H{"message": "ok"})
}
