package articlesSearchRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operator"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/result"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/tokenchar"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/domain/entity/article"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strings"
	"time"
)

const Index = "article"

var logger = logging.GetLogger().With(zap.String("prefix", "[ES]"))

type repository struct {
	client *elasticsearch.TypedClient
}

func New(client *elasticsearch.TypedClient) IRepository {
	return &repository{client}
}

func (r *repository) Setup(ctx context.Context) {
	// 1. Создать индекс публициста
	if r.Exists(ctx) {
		logger.Info(fmt.Sprintf("Индекс %s уже существует, удаляем", Index))
		r.Delete(ctx)
	}

	if err := r.Create(ctx); err != nil {
		logger.Fatal(fmt.Sprintf("Проблема с индексом %s", Index), zap.Error(err))
	}
	// 2. Закинуть дефолтные значения
	articles := []article.EsArticleDBO{
		{
			Name:        "In Israel, Ron DeSantis Promotes His Foreign Policy Credentials",
			Description: "The Florida governor, a likely contender for the Republican presidential nomination, stressed his strong interest in the country’s affairs, an issue that Donald J. Trump once made his own.",
			URL:         "https://www.nytimes.com/2023/04/27/world/middleeast/israel-ron-desantis.html",
			Address: article.AddressES{
				Coords:  [2]float64{40.756133, -73.990322},
				Country: "US",
				City:    "Florida",
			},
			Publisher: article.PublisherES{
				Name: "The New York Times",
				Address: article.AddressES{
					Coords:  [2]float64{40.756133, -73.990322},
					Country: "US",
					City:    "New York",
				},
			},
			Categories: []string{"politics", "economics"},
			People: []article.PersonES{
				{FullName: "Ron DeSantis"},
				{FullName: "Donald J. Trump"},
			},
			Links:         []string{},
			DatePublished: time.Now(),
		},
	}
	for _, a := range articles {
		if ok := r.IndexArticle(ctx, &a); !ok {
			logger.Fatal("Не записали данные в " + Index)
		}
	}

	// 3. Тест поиска
	time.Sleep(2 * time.Second)
	f := dto.GrandFilterRequest{
		FilterStrings: []dto.SearchString{
			{Search: "Ron DeSantis", IsExact: true},
		},
		FilterPeople:     []dto.SearchPeople{},
		FilterPublishers: []dto.SearchPublishers{{Name: "The New York Times"}},
		FilterCountry:    []dto.SearchCountry{{Country: "US"}},
		FilterTime:       dto.SearchTime{},
	}

	if p, err := r.FindArticles(ctx, f); err != nil {
		logger.Fatal("Ошибка во время поиска...", zap.Error(err))
	} else if len(p) == 0 {
		logger.Fatal("Не нашли тестовые данные...")
	} else {
		logger.Info(fmt.Sprintf("Нашли %d статей", len(p)))
	}
}

func (r *repository) Exists(ctx context.Context) bool {
	if resp, err := r.client.Indices.Exists(Index).Perform(ctx); err != nil {
		logger.Error("Не смогли узнать о существовании индекса", zap.Error(err))
		return false
	} else {
		return resp.StatusCode == http.StatusOK
	}
}

func (r *repository) Delete(ctx context.Context) {
	if _, err := r.client.Indices.Delete(Index).Do(ctx); err != nil {
		logger.Error("Не удалили индекс "+Index, zap.Error(err))
	}
}

func (r *repository) Create(ctx context.Context) error {
	articlesTokenizer := types.NGramTokenizer{
		MinGram:    2,
		MaxGram:    20,
		TokenChars: []tokenchar.TokenChar{tokenchar.Letter, tokenchar.Digit},
		Type:       "ngram",
	}

	addressMapping := &types.ObjectProperty{
		Properties: map[string]types.Property{
			"country": types.NewKeywordProperty(),
			"city":    types.NewKeywordProperty(),
			"coords":  types.NewGeoPointProperty(),
		},
		Type: "object",
	}

	res, err := r.client.Indices.Create(Index).
		Request(&create.Request{
			Settings: &types.IndexSettings{
				Analysis: &types.IndexSettingsAnalysis{
					Tokenizer: map[string]types.Tokenizer{
						"article_tokenizer": articlesTokenizer,
					},
					Analyzer: map[string]types.Analyzer{
						"article_analyzer": types.CustomAnalyzer{
							Tokenizer: "article_tokenizer",
							Filter:    []string{types.NewLowercaseTokenFilter().Type},
						},
						"article_search_analyzer": types.NewStandardAnalyzer(),
					},
				},
				MaxNgramDiff: lib.PointerFrom(20),
			},
			Mappings: &types.TypeMapping{
				Properties: map[string]types.Property{
					"name": &types.TextProperty{
						Analyzer:       lib.PointerFrom("article_analyzer"),
						SearchAnalyzer: lib.PointerFrom("article_search_analyzer"),
						Type:           "text",
						Index:          lib.PointerFrom(true),
					},
					"description": &types.TextProperty{
						Analyzer:       lib.PointerFrom("article_analyzer"),
						SearchAnalyzer: lib.PointerFrom("article_search_analyzer"),
						Type:           "text",
						Index:          lib.PointerFrom(true),
					},
					"URL":        types.NewKeywordProperty(),
					"categories": types.NewKeywordProperty(),
					"address":    addressMapping,
					"publisher": &types.ObjectProperty{
						Properties: map[string]types.Property{
							"name":    types.NewKeywordProperty(),
							"address": addressMapping,
						},
						Type: "object",
					},
					"people": &types.ObjectProperty{
						Properties: map[string]types.Property{
							"fullName": types.NewKeywordProperty(),
						},
						Type: "object",
					},
					"links":         types.NewKeywordProperty(),
					"datePublished": types.NewDateProperty(),
				},
			},
		}).
		Do(ctx)

	if err != nil {
		logger.Error("Не создали индекс " + Index)
	} else {
		log.Println(res)
	}

	return err
}

func (r *repository) FindArticles(ctx context.Context, f dto.GrandFilterRequest) ([]*article.EsArticleDBO, error) {

	// 1. Поисковые строки
	var must []types.Query
	for _, filterString := range f.FilterStrings {
		q := &types.Query{Bool: &types.BoolQuery{}}
		s := strings.ToLower(filterString.Search)

		if filterString.IsExact {
			q.Bool.Should = []types.Query{
				{Match: map[string]types.MatchQuery{"name": {Query: s, Operator: &operator.And}}},
				{Match: map[string]types.MatchQuery{"description": {Query: s, Operator: &operator.And}}},
			}
		} else {
			q.Bool.Should = []types.Query{
				{Fuzzy: map[string]types.FuzzyQuery{"name": {Value: s}}},
				{Fuzzy: map[string]types.FuzzyQuery{"description": {Value: s}}},
			}
		}
		must = append(must, *q)
	}

	// 2. Страна
	countryMust := types.Query{Bool: &types.BoolQuery{}}
	for _, country := range f.FilterCountry {
		q := types.Query{Match: map[string]types.MatchQuery{
			"address.country": {Query: country.Country}},
		}
		countryMust.Bool.Should = append(countryMust.Bool.Should, q)
	}
	if len(f.FilterCountry) > 0 {
		must = append(must, countryMust)
	}

	// 3. Издание
	pubMust := types.Query{Bool: &types.BoolQuery{}}
	for _, p := range f.FilterPublishers {
		q := types.Query{Match: map[string]types.MatchQuery{
			"publisher.name": {Query: p.Name},
		}}
		pubMust.Bool.Should = append(pubMust.Bool.Should, q)
	}
	if len(f.FilterPublishers) > 0 {
		must = append(must, pubMust)
	}

	// 4. Люди
	peopleMust := types.Query{Bool: &types.BoolQuery{}}
	for _, person := range f.FilterPeople {
		q := types.Query{Match: map[string]types.MatchQuery{
			"people.fullName": {Query: person.Name},
		}}
		peopleMust.Bool.Should = append(pubMust.Bool.Should, q)
	}
	if len(f.FilterPeople) > 0 {
		must = append(must, peopleMust)
	}

	// 5. Время f.FilterTime

	req := &search.Request{
		Size: lib.PointerFrom(100),
		Query: &types.Query{
			Bool: &types.BoolQuery{Must: must},
		},
	}

	resp, err := r.client.Search().
		Index(Index).
		Request(req).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("По запросу grandFilter нашли %d", resp.Hits.Total.Value))

	var p []*article.EsArticleDBO
	for _, hit := range resp.Hits.Hits {
		var res *article.EsArticleDBO
		if err := json.Unmarshal(hit.Source_, &res); err != nil {
			return nil, err
		}
		p = append(p, res)
	}

	return p, nil
}

func (r *repository) IndexArticle(ctx context.Context, a *article.EsArticleDBO) bool {
	res, err := r.client.Index(Index).
		Request(a).
		Do(ctx)

	if err != nil {
		logger.Error(fmt.Sprintf("Не записали данные в %s", Index), zap.Error(err))
	} else {
		logger.Info(fmt.Sprintf("Добавили %s в ES[%s] с id=[%s]", a.Name, Index, res.Id_))
	}

	return res.Result == result.Created
}
