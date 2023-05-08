package searcher

import (
	"github.com/gin-gonic/gin"
	"github.com/mskKote/prospero_backend/pkg/handlers"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"net/http"
)

type handler struct {
	logger logging.Logger
}

func New(logger *logging.Logger) handlers.Handler {
	return &handler{*logger}
}

func (h *handler) Register(g *gin.RouterGroup) {
	g.POST("/grandFilter", h.GrandFilter)
}

func (h *handler) GrandFilter(g *gin.Context) {
	h.logger.Info("Вызвали GrandFilter")
	g.JSON(http.StatusOK, gin.H{"message": "ok"})
}
