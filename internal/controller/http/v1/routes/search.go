package routes

import (
	"github.com/gin-gonic/gin"
)

const (
	searchURL = "/grandFilter"
)

type ISearchUsecase interface {
	GrandFilter(g *gin.Context)
}

type SearchHandlers struct {
	search ISearchUsecase
}

func NewSearchRoute(search ISearchUsecase) *SearchHandlers {
	return &SearchHandlers{search}
}

func (h *SearchHandlers) Register(g *gin.RouterGroup) {
	g.POST(searchURL, h.search.GrandFilter)
}
