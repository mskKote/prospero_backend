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

type ISearchUsecase interface {
	GrandFilter(g *gin.Context)
	SearchPublisherWithHints(c *gin.Context)
	SearchDefaultPublisherWithHints(c *gin.Context)
	SearchLanguages(c *gin.Context)
	SearchCategoriesWithHints(c *gin.Context)
	SearchPeopleWithHints(c *gin.Context)
}

func RegisterSearchRoutes(g *gin.RouterGroup, s ISearchUsecase) {
	g.POST(searchURL, s.GrandFilter)
	g.GET(searchPublisherURL, s.SearchPublisherWithHints)
	g.GET(searchDefaultPublisherURL, s.SearchDefaultPublisherWithHints)
	g.GET(searchLanguagesURL, s.SearchLanguages)
	g.GET(searchCategoriesWithHintsURL, s.SearchCategoriesWithHints)
	g.GET(searchPeopleWithHintsURL, s.SearchPeopleWithHints)
}
