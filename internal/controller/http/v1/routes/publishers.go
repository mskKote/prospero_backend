package routes

import "github.com/gin-gonic/gin"

const (
	createPublisherURL = "/addPublisher"
	readPublishersURL  = "/getPublishers"
	updatePublisherURL = "/updatePublisher"
	deletePublisherURL = "/removePublisher"
)

type IPublishersUseCase interface {
	CreatePublisher(c *gin.Context)
	ReadPublishers(c *gin.Context)
	UpdatePublisher(c *gin.Context)
	DeletePublisher(c *gin.Context)
}

func RegisterPublishersRoutes(g *gin.RouterGroup, p IPublishersUseCase) {
	g.POST(createPublisherURL, p.CreatePublisher)
	g.GET(readPublishersURL, p.ReadPublishers)
	g.PUT(updatePublisherURL, p.UpdatePublisher)
	g.DELETE(deletePublisherURL, p.DeletePublisher)
}
