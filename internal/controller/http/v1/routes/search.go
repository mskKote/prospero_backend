package routes

import (
	"github.com/gin-gonic/gin"
)

const (
	searchURL          = "/grandFilter/:search"
	searchPublisherURL = "/searchPublisherWithHints/:search"
)

type ISearchUsecase interface {
	GrandFilter(g *gin.Context)
	SearchPublisherWithHints(c *gin.Context)
}

func RegisterSearchRoutes(g *gin.RouterGroup, s ISearchUsecase) {
	g.POST(searchURL, s.GrandFilter)
	g.GET(searchPublisherURL, s.SearchPublisherWithHints)
}
