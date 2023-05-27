package RSS

import "github.com/mmcdole/gofeed"

type IParserUsecase interface {
	// Startup запускает постоянный RSS парсер
	Startup()
	// ParseRSS достаёт контент по источнику
	ParseRSS(src string) *gofeed.Feed
}
