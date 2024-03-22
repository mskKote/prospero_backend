package routes

import (
	"github.com/gin-gonic/gin"
)

const (
	searchURL                    = "/grandFilter"
	searchPublisherURL           = "/searchPublisherWithHints/:search"
	searchDefaultPublisherURL    = "/searchPublisherWithHints/"
	searchLanguagesURL           = "/searchLanguages"
	searchCategoriesWithHintsURL = "/searchCategoryWithHints"
	searchPeopleWithHintsURL     = "/searchPeopleWithHints"
)

type ISearchUseCase interface {
	GrandFilter(g *gin.Context)
	SearchPublisherWithHints(c *gin.Context)
	SearchDefaultPublisherWithHints(c *gin.Context)
	SearchLanguages(c *gin.Context)
	SearchCategoriesWithHints(c *gin.Context)
	SearchPeopleWithHints(c *gin.Context)
}

func RegisterSearchRoutes(g *gin.RouterGroup, s ISearchUseCase) {
	g.POST(searchURL, s.GrandFilter)
	g.POST(searchPublisherURL, s.SearchPublisherWithHints)
	g.POST(searchDefaultPublisherURL, s.SearchDefaultPublisherWithHints)
	g.POST(searchLanguagesURL, s.SearchLanguages)
	g.POST(searchCategoriesWithHintsURL, s.SearchCategoriesWithHints)
	g.POST(searchPeopleWithHintsURL, s.SearchPeopleWithHints)
}
