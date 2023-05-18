package routes

import (
	"github.com/gin-gonic/gin"
)

const (
	searchURL = "/grandFilter/:search"
)

type ISearchUsecase interface {
	GrandFilter(g *gin.Context)
}

func RegisterSearchRoutes(g *gin.RouterGroup, search ISearchUsecase) {
	g.POST(searchURL, search.GrandFilter)
}
