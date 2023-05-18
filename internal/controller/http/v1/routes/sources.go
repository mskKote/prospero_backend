package routes

import "github.com/gin-gonic/gin"

const (
	createSourceURL        = "/RSS/addSource"
	readSourcesURL         = "/RSS/getSources"
	readEnrichedSourcesURL = "/RSS/getEnrichedSources"
	updateSourceURL        = "/RSS/updateSource"
	deleteSourceURL        = "/RSS/removeSource"
)

type ISourcesUsecase interface {
	CreateSourceRSS(c *gin.Context)
	ReadSourcesRSS(c *gin.Context)
	ReadSourcesRSSWithPublishers(c *gin.Context)
	UpdateSourceRSS(c *gin.Context)
	DeleteSourceRSS(c *gin.Context)
}

func RegisterSourcesRoutes(g *gin.RouterGroup, sources ISourcesUsecase) {
	g.POST(createSourceURL, sources.CreateSourceRSS)
	g.GET(readSourcesURL, sources.ReadSourcesRSS)
	g.GET(readEnrichedSourcesURL, sources.ReadSourcesRSSWithPublishers)
	g.PUT(updateSourceURL, sources.UpdateSourceRSS)
	g.DELETE(deleteSourceURL, sources.DeleteSourceRSS)
}
