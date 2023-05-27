package routes

import (
	"github.com/gin-gonic/gin"
)

const (
	searchURL                 = "/grandFilter"
	searchPublisherURL        = "/searchPublisherWithHints/:search"
	searchDefaultPublisherURL = "/searchPublisherWithHints/"
)

type ISearchUsecase interface {
	GrandFilter(g *gin.Context)
	SearchPublisherWithHints(c *gin.Context)
	SearchDefaultPublisherWithHints(c *gin.Context)
}

func RegisterSearchRoutes(g *gin.RouterGroup, s ISearchUsecase) {
	g.POST(searchURL, s.GrandFilter)
	g.GET(searchPublisherURL, s.SearchPublisherWithHints)
	g.GET(searchDefaultPublisherURL, s.SearchDefaultPublisherWithHints)
}
