package routes

import "github.com/gin-gonic/gin"

const (
	createSourceURL        = "/RSS/addSource"
	readSourcesURL         = "/RSS/getSources"
	readEnrichedSourcesURL = "/RSS/getEnrichedSources"
	updateSourceURL        = "/RSS/updateSource"
	deleteSourceURL        = "/RSS/removeSource"
	harvest                = "/RSS/harvest"
	addSourceAndPublisher  = "/addSourceAndPublisher"
)

type ISourcesUsecase interface {
	CreateSourceRSS(c *gin.Context)
	ReadSourcesRSS(c *gin.Context)
	ReadSourcesRSSWithPublishers(c *gin.Context)
	UpdateSourceRSS(c *gin.Context)
	DeleteSourceRSS(c *gin.Context)
	Harvest(c *gin.Context)
	AddSourceAndPublisher(c *gin.Context)
}

func RegisterSourcesRoutes(g *gin.RouterGroup, sources ISourcesUsecase) {
	g.POST(addSourceAndPublisher, sources.AddSourceAndPublisher)
	g.POST(createSourceURL, sources.CreateSourceRSS)
	g.POST(harvest, sources.Harvest)
	g.GET(readSourcesURL, sources.ReadSourcesRSS)
	g.GET(readEnrichedSourcesURL, sources.ReadSourcesRSSWithPublishers)
	g.PUT(updateSourceURL, sources.UpdateSourceRSS)
	g.DELETE(deleteSourceURL, sources.DeleteSourceRSS)
}
