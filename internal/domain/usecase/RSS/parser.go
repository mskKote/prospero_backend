package RSS

import (
	"encoding/json"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
	"strings"
)

var (
	logger = logging.GetLogger()
	parser = gofeed.NewParser()
)

/*
TODO: PROS-29 подключаю RSS
2 собираю источники раз в N времени
3 база данных источников
4 сохранение статей
*/

// Service - зависимые сервисы (ElasticSearch)
type services interface {
	// Startup запускает постоянный RSS парсер
	Startup()
	// AddSourceRSS добавляет RSS ссылку
	AddSourceRSS(src string)
	// ParseRSS достаёт контент по источнику
	ParseRSS(src string)
}

// Usecase использование сервисов
type Usecase struct {
	services
}

var sources = []string{
	"https://rss.nytimes.com/services/xml/rss/nyt/World.xml",
}

func (u *Usecase) AddSource(src string) {
	sources = append(sources, src)
}

func (u *Usecase) Startup() {
	for i, source := range sources {
		logger.Info("Парсим источинк #" + string(rune(i)))
		u.ParseRSS(source)
	}
}

func (u *Usecase) ParseRSS(src string) {
	feed, _ := parser.ParseURL(src)
	article, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		logger.Error("Не смогли распарсить источник", zap.Error(err))
	} else {
		for _, s := range strings.Split(string(article), "\n") {
			fmt.Println(s)
		}
	}
}
