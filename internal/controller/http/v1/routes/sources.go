package routes

import "github.com/gin-gonic/gin"

const (
	addSourceURL     = "/addSourceRSS"
	getSourcesURL    = "/getSourcesURL"
	deleteSourcesURL = "/deleteSourcesURL"
)

type ISourcesUsecase interface {
	AddSourceRSS(c *gin.Context)
	GetSourcesRSS(c *gin.Context)
	DeleteSourcesRSS(c *gin.Context)
}

type SourcesHandlers struct {
	sources ISourcesUsecase
}

func NewSourcesRoutes(sources ISourcesUsecase) *SourcesHandlers {
	return &SourcesHandlers{sources}
}

func (h *SourcesHandlers) RegisterSources(g *gin.RouterGroup) {
	g.POST(addSourceURL, h.sources.AddSourceRSS)
	g.GET(getSourcesURL, h.sources.GetSourcesRSS)
	g.DELETE(deleteSourcesURL, h.sources.DeleteSourcesRSS)
}
