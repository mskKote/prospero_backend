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
	"github.com/mskKote/prospero_backend/internal/domain/entity/source"
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
		partPotential := make(chan int)
		for srcI, src := range sources {
			go func(_srcI int, _src *source.RSS) {
				logger.Info(fmt.Sprintf("Парсим источник #%d: %s", i*batch+_srcI+1, _src.RssURL))

				feed := s.ParseRSS(_src.RssURL)
				//u.logFeed(feed)
				feedPotential := s.analyseFeed(feed)
				// сохранить новости в ES
				s.indexFeed(ctx, _src.Publisher.ToDTO(), feed)
				partPotential <- feedPotential
			}(srcI, src)
		}
		// Ждём чтение партии
		for range sources {
			potential += <-partPotential
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
	potentialChan := make(chan int)
	// TODO: достать слова с большой буквы
	// TODO: обгатить новости в NameAPI и 2gis
	for _, item := range feed.Items {
		go func(_item *gofeed.Item) {
			itemPotential := 0
			for _, s := range strings.Fields(_item.Description) {
				if matched, err := regexp.MatchString(`^[A-Z|А-Я][a-z|а-я]+`, s); err != nil {
					logger.Error("Проблема в regexp", zap.Error(err))
				} else if matched {
					//logger.Info("POTENTIAL NAME/TOPONYM : " + s)
					itemPotential++
				}
			}
			potentialChan <- itemPotential
		}(item)
	}
	for range feed.Items {
		potential += <-potentialChan
	}
	return potential
}

func (s *service) indexFeed(ctx context.Context, p *publisher.DTO, feed *gofeed.Feed) {
	itemsChan := make(chan bool)

	for _, item := range feed.Items {
		go func(_item *gofeed.Item) {
			var people []article.PersonES
			for _, author := range _item.Authors {
				people = append(people, article.PersonES{FullName: author.Name})
			}
			language := strings.ToLower(strings.Split(feed.Language, "-")[0])

			articleDBO := &article.EsArticleDBO{
				Name:        _item.Title,
				Description: _item.Description,
				URL:         _item.Link,
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
				Categories:    _item.Categories,
				People:        people,
				Links:         _item.Links,
				DatePublished: _item.PublishedParsed,
				Language:      language,
			}
			if ok := s.elastic.IndexArticle(ctx, articleDBO); !ok {
				logger.ErrorContext(ctx, "Ошибка добавления статьи")
			}

			for _, category := range _item.Categories {
				categoryDBO := &article.CategoryES{Name: category}

				s.elastic.IndexCategory(ctx, categoryDBO)
			}
			for _, person := range people {
				s.elastic.IndexPeople(ctx, &person)
			}
			itemsChan <- true
		}(item)
	}
	for range feed.Items {
		<-itemsChan
	}
}

func (s *service) ParseRSS(src string) *gofeed.Feed {
	feed, err := parser.ParseURL(src)
	if err != nil {
		logger.Error("Ошибка парса", zap.Error(err))
	}
	return feed
}

func (s *service) FindWithGrandFilter(ctx context.Context, p dto.GrandFilterRequest, size int) ([]*article.EsArticleDBO, int64, error) {
	tracer := tracing.TracerFromContext(ctx)
	ctxWithSpan, span := tracer.Start(ctx, "ElasticSearch")
	logger.InfoContext(ctxWithSpan, "Создали Span")
	span.SetAttributes(attribute.String("[articleSERVICE]", "Идём в ElasticSearch"))
	defer span.End()

	return s.elastic.FindArticles(ctxWithSpan, p, size)
}

func (s *service) FindAllLanguages(ctx context.Context) ([]*article.LanguageES, error) {
	tracer := tracing.TracerFromContext(ctx)
	ctxWithSpan, span := tracer.Start(ctx, "ElasticSearch")
	logger.InfoContext(ctxWithSpan, "Создали Span")
	span.SetAttributes(attribute.String("[articleSERVICE]", "Идём в ElasticSearch"))
	defer span.End()

	return s.elastic.FindLanguages(ctx)
}

func (s *service) FindCategory(ctx context.Context, cat string) ([]*article.CategoryES, error) {
	tracer := tracing.TracerFromContext(ctx)
	ctxWithSpan, span := tracer.Start(ctx, "ElasticSearch")
	logger.InfoContext(ctxWithSpan, "Создали Span")
	span.SetAttributes(attribute.String("[articleSERVICE]", "Идём в ElasticSearch"))
	defer span.End()

	return s.elastic.FindCategory(ctx, cat)
}

func (s *service) FindPeople(ctx context.Context, name string) ([]*article.PersonES, error) {
	return s.elastic.FindPeople(ctx, name)
}

func New(
	sources sourcesRepository.IRepository,
	elastic articlesSearchRepository.IRepository) IArticleService {
	return &service{sources, elastic}
}
