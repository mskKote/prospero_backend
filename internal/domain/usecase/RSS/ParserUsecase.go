package RSS

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/mmcdole/gofeed"
	customMetrics "github.com/mskKote/prospero_backend/internal/adapters/metrics"
	"github.com/mskKote/prospero_backend/internal/domain/entity/article"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
	"github.com/mskKote/prospero_backend/internal/domain/service/articleService"
	"github.com/mskKote/prospero_backend/internal/domain/service/sourcesService"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"github.com/mskKote/prospero_backend/pkg/metrics"

	"go.uber.org/zap"
	"regexp"
	"strings"
	"time"
)

var (
	logger = logging.GetLogger()
	cfg    = config.GetConfig()
	parser = gofeed.NewParser()
)

// usecase использование сервисов
type usecase struct {
	sources  sourcesService.ISourceService
	articles articleService.IArticleService
}

func New(
	s sourcesService.ISourceService,
	a articleService.IArticleService) IParserUsecase {
	return &usecase{
		sources:  s,
		articles: a,
	}
}

// Startup - запускает cron job
// для парсинга источников из postgres.
// Время работы определяется в app.yml
func (u *usecase) Startup() {
	s := gocron.NewScheduler(time.UTC)
	logger.Info("Парсим каждые " + cfg.CronSourcesRSS)

	if _, err := s.Cron(cfg.CronSourcesRSS).Do(u.parseJob); err != nil {
		logger.Fatal("Не стартовали CRON RSS", zap.Error(err))
	}

	// При миграции всё удаляется, достаю данные
	if cfg.MigrateElastic {
		logger.Info("[MIGRATION] Закинуть данные в ELASTIC сразу")
		u.parseJob()
	}

	s.StartAsync()
}

func (u *usecase) parseJob() {
	start := time.Now()
	ctx := context.Background()
	count, err := u.sources.Count(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "[POSTGRES] Не посчитали источники", zap.Error(err))
		return
	}
	const batch = 8
	parts := int(count / batch)

	potential := 0

	for i := 0; i <= parts; i++ {
		sources, err := u.sources.FindAllWithPublisher(ctx, i, batch)
		if err != nil {
			logger.ErrorContext(ctx,
				fmt.Sprintf("Не прочитали партию источников {%d/%d}", i+1, parts+1),
				zap.Error(err))
			continue
		}
		logger.InfoContext(ctx,
			fmt.Sprintf("Прочитали партию источников {%d/%d}", i+1, parts+1))

		// Читаем партию источников
		for srcI, src := range sources {
			logger.Info(fmt.Sprintf("Парсим источник #%d: %s", i*batch+srcI+1, src.RssURL))

			feed := u.ParseRSS(src.RssURL)
			//u.logFeed(feed)
			potential += u.analyseFeed(feed)
			// сохранить новости в ES
			u.indexFeed(ctx, src.Publisher.ToDTO(), feed)
		}
	}

	logger.InfoContext(ctx, fmt.Sprintf("Потенциальные топонимы/имена %d", potential))

	elapsed := time.Since(start)

	metrics.ObserveSummaryMetric(customMetrics.MetricRssObtainName, elapsed.Seconds())
}

func (u *usecase) ParseRSS(src string) *gofeed.Feed {
	feed, err := parser.ParseURL(src)
	if err != nil {
		// TODO: обработать ошибку парса
	}
	return feed
}

func (u *usecase) logFeed(feed *gofeed.Feed) {
	jsonFeed, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		logger.Error("Не смогли распарсить источник", zap.Error(err))
		return
	}

	for _, s := range strings.Split(string(jsonFeed), "\n") {
		fmt.Println(s)
	}
}

func (u *usecase) analyseFeed(feed *gofeed.Feed) int {
	potential := 0
	// TODO: достать слова с большой буквы
	// TODO: обгатить новости в NameAPI и 2gis
	for _, item := range feed.Items {
		for _, s := range strings.Fields(item.Description) {
			if matched, err := regexp.MatchString(`^[A-Z|А-Я][a-z|а-я]+`, s); err != nil {
				logger.Error("Проблема в regexp", zap.Error(err))
			} else if matched {
				//logger.Info("POTENTIAL NAME/TOPONYM : " + s)
				potential++
			}
		}
	}
	return potential
}

func (u *usecase) indexFeed(ctx context.Context, p *publisher.DTO, feed *gofeed.Feed) {
	for _, item := range feed.Items {
		var people []article.PersonES
		for _, author := range item.Authors {
			people = append(people, article.PersonES{FullName: author.Name})
		}

		articleDBO := &article.EsArticleDBO{
			Name:        item.Title,
			Description: item.Description,
			URL:         item.Link,
			// TODO Address PlaceAPI
			Address: article.AddressES{
				Coords:  [2]float64{p.Latitude, p.Longitude},
				Country: p.Country,
				City:    p.City,
			},
			Publisher: article.PublisherES{
				Name: p.Name,
				Address: article.AddressES{
					Coords:  [2]float64{p.Latitude, p.Longitude},
					Country: p.Country,
					City:    p.City,
				},
			},
			Categories: item.Categories,
			// TODO People по NameAPI
			People:        people,
			Links:         item.Links,
			DatePublished: item.PublishedParsed,
		}

		if ok := u.articles.AddArticle(ctx, articleDBO); !ok {
			logger.FatalContext(ctx, "Ошибка добавления статьи")
		} else {
			//logger.InfoContext(ctx, "Добавили статью")
		}
	}
}
