package RSS

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/mmcdole/gofeed"
	"github.com/mskKote/prospero_backend/internal/domain/service/sourcesService"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
	"log"
	"regexp"
	"strings"
	"time"
)

var (
	logger = logging.GetLogger()
	cfg    = config.GetConfig()
	parser = gofeed.NewParser()
)

// Service - зависимые сервисы
type services interface {
	// Startup запускает постоянный RSS парсер
	Startup()
	// ParseRSS достаёт контент по источнику
	ParseRSS(src string)
}

// Usecase использование сервисов
type Usecase struct {
	services
	sources sourcesService.ISourceService
}

func New(s sourcesService.ISourceService) *Usecase {
	return &Usecase{sources: s}
}

/*
TODO: [PROS-29] подключаю RSS
4 сохранение статей
*/

// Startup - запускает cron job
// для парсинга источников из postgres.
// Время работы определяется в app.yml
func (u *Usecase) Startup() {
	s := gocron.NewScheduler(time.UTC)
	logger.Info("Парсим каждые " + cfg.CronSourcesRSS)

	_, err := s.Cron(cfg.CronSourcesRSS).Do(u.parseJob)
	if err != nil {
		logger.Fatal("Не стартовали CRON RSS", zap.Error(err))
	}
	s.StartAsync()
}

func (u *Usecase) parseJob() {
	ctx := context.Background()
	count, err := u.sources.Count(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "[POSTGRES] Не посчитали источники", zap.Error(err))
		return
	}
	const batch = 8
	parts := int(count / batch)

	for i := 0; i <= parts; i++ {
		sources, err := u.sources.FindAll(ctx, i, batch)
		if err != nil {
			logger.ErrorContext(ctx,
				fmt.Sprintf("Не прочитали партию источников {%d/%d}", i+1, parts+1),
				zap.Error(err))
			continue
		}
		logger.InfoContext(ctx,
			fmt.Sprintf("Прочитали партию источников {%d/%d}", i+1, parts+1))

		// Читаем партию источников
		for srcI, source := range sources {
			logger.Info(fmt.Sprintf("Парсим источник #%d: %s", i*batch+srcI+1, source.RssURL))
			feed := u.ParseRSS(source.RssURL)
			//u.logFeed(feed)
			u.analyseFeed(feed)
			return
			// TODO: сохранить новости в ES
		}
	}
}

func (u *Usecase) ParseRSS(src string) *gofeed.Feed {
	feed, err := parser.ParseURL(src)
	if err != nil {

	}
	return feed
}

func (u *Usecase) logFeed(feed *gofeed.Feed) {
	article, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		logger.Error("Не смогли распарсить источник", zap.Error(err))
		return
	}

	for _, s := range strings.Split(string(article), "\n") {
		fmt.Println(s)
	}
}

func (u *Usecase) analyseFeed(feed *gofeed.Feed) {
	// TODO: достать слова с большой буквы
	// TODO: обгатить новости в NameAPI и 2gis
	for _, item := range feed.Items {
		article := item.Title + " " + item.Description
		log.Println(article)
		for _, s := range strings.Fields(article) {
			matched, err := regexp.MatchString(`^[A-Z|А-Я][a-z|а-я]+`, s)
			if err != nil {
				logger.Error("Проблема в regexp", zap.Error(err))
			}
			if matched {
				logger.Info("POTENTIAL NAME/TOPONYM : " + s)
			}
		}

		//item.Categories
		//item.Authors
		//item.Published
	}
}
