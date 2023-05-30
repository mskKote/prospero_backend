package articleService

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mmcdole/gofeed"
	"github.com/mskKote/prospero_backend/internal/adapters/db/elastic/articlesSearchRepository"
	"github.com/mskKote/prospero_backend/internal/adapters/db/postgres/sourcesRepository"
	customMetrics "github.com/mskKote/prospero_backend/internal/adapters/metrics"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/domain/entity/article"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"github.com/mskKote/prospero_backend/pkg/metrics"
	"github.com/mskKote/prospero_backend/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"regexp"
	"strings"
	"time"
)

var (
	logger = logging.GetLogger()
	parser = gofeed.NewParser()
)

type service struct {
	sources sourcesRepository.IRepository
	elastic articlesSearchRepository.IRepository
}

func (s *service) ParseAllOnce(ctx context.Context) error {
	start := time.Now()
	count, err := s.sources.Count(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "[POSTGRES] Не посчитали источники", zap.Error(err))
		return err
	}
	const batch = 8
	parts := int(count / batch)

	potential := 0

	for i := 0; i <= parts; i++ {
		sources, err := s.sources.FindAllWithPublishers(ctx, i*batch, batch)
		if err != nil {
			logger.ErrorContext(ctx,
				fmt.Sprintf("Не прочитали партию источников {%d/%d}", i+1, parts+1),
				zap.Error(err))
			return err
		}
		logger.InfoContext(ctx,
			fmt.Sprintf("Прочитали партию источников {%d/%d}", i+1, parts+1))

		// Читаем партию источников
		for srcI, src := range sources {
			logger.Info(fmt.Sprintf("Парсим источник #%d: %s", i*batch+srcI+1, src.RssURL))

			feed := s.ParseRSS(src.RssURL)
			//u.logFeed(feed)
			potential += s.analyseFeed(feed)
			// сохранить новости в ES
			s.indexFeed(ctx, src.Publisher.ToDTO(), feed)
		}
	}

	logger.InfoContext(ctx, fmt.Sprintf("Потенциальные топонимы/имена %d", potential))

	elapsed := time.Since(start)
	metrics.ObserveSummaryMetric(customMetrics.MetricRssObtainName, elapsed.Seconds())
	return nil
}

func (s *service) logFeed(feed *gofeed.Feed) {
	jsonFeed, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		logger.Error("Не смогли распарсить источник", zap.Error(err))
		return
	}

	for _, s := range strings.Split(string(jsonFeed), "\n") {
		fmt.Println(s)
	}
}

func (s *service) analyseFeed(feed *gofeed.Feed) int {
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

func (s *service) indexFeed(ctx context.Context, p *publisher.DTO, feed *gofeed.Feed) {
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

		if ok := s.AddArticle(ctx, articleDBO); !ok {
			logger.FatalContext(ctx, "Ошибка добавления статьи")
		} else {
			//logger.InfoContext(ctx, "Добавили статью")
		}
	}
}

func (s *service) ParseRSS(src string) *gofeed.Feed {
	feed, err := parser.ParseURL(src)
	if err != nil {
		logger.Error("Ошибка парса", zap.Error(err))
	}
	return feed
}

func (s *service) AddArticle(ctx context.Context, dto *article.EsArticleDBO) bool {
	return s.elastic.IndexArticle(ctx, dto)
}

func (s *service) FindWithGrandFilter(ctx context.Context, p dto.GrandFilterRequest) ([]*article.EsArticleDBO, error) {
	tracer := tracing.TracerFromContext(ctx)
	ctxWithSpan, span := tracer.Start(ctx, "ElasticSearch")
	logger.InfoContext(ctxWithSpan, "Создали Span")
	span.SetAttributes(attribute.String("[articleSERVICE]", "Идём в ElasticSearch"))
	defer span.End()

	return s.elastic.FindArticles(ctxWithSpan, p)
}

func New(
	sources sourcesRepository.IRepository,
	elastic articlesSearchRepository.IRepository) IArticleService {
	return &service{sources, elastic}
}
