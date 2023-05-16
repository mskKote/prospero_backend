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

type SearchHandlers struct {
	search ISearchUsecase
}

func NewSearchRoutes(search ISearchUsecase) *SearchHandlers {
	return &SearchHandlers{search}
}

func (h *SearchHandlers) RegisterSearch(g *gin.RouterGroup) {
	g.POST(searchURL, h.search.GrandFilter)
}
