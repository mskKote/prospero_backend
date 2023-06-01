package articlesSearchRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/fieldsortnumerictype"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operator"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/optype"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/result"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/tokenchar"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/domain/entity/article"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"hash/fnv"
	"log"
	"net/http"
	"strings"
)

const (
	ArticleIndex  = "article"
	CategoryIndex = "category"
	PeopleIndex   = "people"
)

var (
	logger  = logging.GetLogger().With(zap.String("prefix", "[ES]"))
	Indices = [...]string{ArticleIndex, CategoryIndex, PeopleIndex}
)

type repository struct {
	client *elasticsearch.TypedClient
}

func New(client *elasticsearch.TypedClient) IRepository {
	return &repository{client}
}

func (r *repository) Setup(ctx context.Context) {
	// 1. Создать индексы
	for _, index := range Indices {
		if r.Exists(ctx, index) {
			logger.Info(fmt.Sprintf("Индекс %s уже существует, удаляем", index))
			r.Delete(ctx, index)
		}
	}

	if err := r.Create(ctx); err != nil {
		logger.Fatal(fmt.Sprintf("Проблема с индексами"), zap.Error(err))
	}
	// 2. Закинуть дефолтные значения
	//articles := []article.EsArticleDBO{
	//	{
	//		Name:        "In Israel, Ron DeSantis Promotes His Foreign Policy Credentials",
	//		Description: "The Florida governor, a likely contender for the Republican presidential nomination, stressed his strong interest in the country’s affairs, an issue that Donald J. Trump once made his own.",
	//		URL:         "https://www.nytimes.com/2023/04/27/world/middleeast/israel-ron-desantis.html",
	//		Address: article.AddressES{
	//			Coords:  [2]float64{40.756133, -73.990322},
	//			Country: "US",
	//			City:    "Florida",
	//		},
	//		Publisher: article.PublisherES{
	//			Name: "The New York Times",
	//			Address: article.AddressES{
	//				Coords:  [2]float64{40.756133, -73.990322},
	//				Country: "US",
	//				City:    "New York",
	//			},
	//		},
	//		Categories: []string{"politics", "economics"},
	//		People: []article.PersonES{
	//			{FullName: "Ron DeSantis"},
	//			{FullName: "Donald J. Trump"},
	//		},
	//		Links:         []string{},
	//		DatePublished: lib.PointerFrom(time.Now()),
	//	},
	//}
	//for _, a := range articles {
	//	if ok := r.IndexArticle(ctx, &a); !ok {
	//		logger.Fatal("Не записали данные в " + ArticleIndex)
	//	}
	//}

	// 3. Тест поиска
	//time.Sleep(2 * time.Second)
	//f := dto.GrandFilterRequest{
	//	FilterStrings: []dto.SearchString{
	//		{Search: "Ron DeSantis", IsExact: true},
	//	},
	//	FilterPeople:     []dto.SearchPeople{},
	//	FilterPublishers: []dto.SearchPublishers{{Name: "The New York Times"}},
	//	FilterCountry:    []dto.SearchCountry{{Country: "US"}},
	//	FilterTime:       dto.SearchTime{},
	//}
	//
	//if p, err := r.FindArticles(ctx, f); err != nil {
	//	logger.Fatal("Ошибка во время поиска...", zap.Error(err))
	//} else if len(p) == 0 {
	//	logger.Fatal("Не нашли тестовые данные...")
	//} else {
	//	logger.Info(fmt.Sprintf("Нашли %d статей", len(p)))
	//}
}

func (r *repository) Exists(ctx context.Context, index string) bool {
	if resp, err := r.client.Indices.Exists(index).Perform(ctx); err != nil {
		logger.Error("Не смогли узнать о существовании индекса", zap.Error(err))
		return false
	} else {
		return resp.StatusCode == http.StatusOK
	}
}

func (r *repository) Delete(ctx context.Context, index string) {
	if _, err := r.client.Indices.Delete(index).Do(ctx); err != nil {
		logger.Error("Не удалили индекс "+index, zap.Error(err))
	}
}

func (r *repository) Create(ctx context.Context) error {
	// ------------------------------------------------- Статьи
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

	res, err := r.client.Indices.Create(ArticleIndex).
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
						"article_search_analyzer": types.NewWhitespaceAnalyzer(),
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
					"language":      types.NewKeywordProperty(),
					"datePublished": types.NewDateProperty(),
				},
			},
		}).
		Do(ctx)

	if err != nil {
		logger.Error("Не создали индекс " + ArticleIndex)
		return err
	} else {
		log.Println(res)
	}

	// ------------------------------------------------- Категории
	categoryTokenizer := types.NGramTokenizer{
		MinGram: 2,
		MaxGram: 20,
		TokenChars: []tokenchar.TokenChar{
			tokenchar.Letter,
			tokenchar.Digit,
			tokenchar.Whitespace,
			tokenchar.Punctuation,
			tokenchar.Symbol},
		Type: "ngram",
	}

	res, err = r.client.Indices.Create(CategoryIndex).
		Request(&create.Request{
			Settings: &types.IndexSettings{
				Analysis: &types.IndexSettingsAnalysis{
					Tokenizer: map[string]types.Tokenizer{
						"category_tokenizer": categoryTokenizer,
					},
					Analyzer: map[string]types.Analyzer{
						"category_analyzer": types.CustomAnalyzer{
							Tokenizer: "category_tokenizer",
							Filter:    []string{types.NewLowercaseTokenFilter().Type},
						},
						"category_search_analyzer": types.NewKeywordAnalyzer(),
					},
				},
				MaxNgramDiff: lib.PointerFrom(20),
			},
			Mappings: &types.TypeMapping{
				Properties: map[string]types.Property{
					"name": &types.TextProperty{
						Analyzer:       lib.PointerFrom("category_analyzer"),
						SearchAnalyzer: lib.PointerFrom("category_search_analyzer"),
						Type:           "text",
						Index:          lib.PointerFrom(true),
					},
				},
			},
		}).
		Do(ctx)

	if err != nil {
		logger.Error("Не создали индекс " + CategoryIndex)
		return err
	} else {
		log.Println(res)
	}

	// ------------------------------------------------- Люди
	peopleTokenizer := types.NGramTokenizer{
		MinGram: 2,
		MaxGram: 20,
		TokenChars: []tokenchar.TokenChar{
			tokenchar.Letter,
			tokenchar.Digit,
			tokenchar.Whitespace,
			tokenchar.Punctuation,
			tokenchar.Symbol},
		Type: "ngram",
	}

	res, err = r.client.Indices.Create(PeopleIndex).
		Request(&create.Request{
			Settings: &types.IndexSettings{
				Analysis: &types.IndexSettingsAnalysis{
					Tokenizer: map[string]types.Tokenizer{
						"people_tokenizer": peopleTokenizer,
					},
					Analyzer: map[string]types.Analyzer{
						"people_analyzer": types.CustomAnalyzer{
							Tokenizer: "people_tokenizer",
							Filter:    []string{types.NewLowercaseTokenFilter().Type},
						},
						"people_search_analyzer": types.NewKeywordAnalyzer(),
					},
				},
				MaxNgramDiff: lib.PointerFrom(20),
			},
			Mappings: &types.TypeMapping{
				Properties: map[string]types.Property{
					"fullName": &types.TextProperty{
						Analyzer:       lib.PointerFrom("people_analyzer"),
						SearchAnalyzer: lib.PointerFrom("people_search_analyzer"),
						Type:           "text",
						Index:          lib.PointerFrom(true),
					},
				},
			},
		}).
		Do(ctx)

	if err != nil {
		logger.Error("Не создали индекс " + CategoryIndex)
		return err
	} else {
		log.Println(res)
	}

	return nil
}

func (r *repository) IndexArticle(ctx context.Context, a *article.EsArticleDBO) bool {
	res, err := r.client.Index(ArticleIndex).
		Request(a).
		Do(ctx)

	if err != nil {
		logger.Error(fmt.Sprintf("Не записали данные в %s", ArticleIndex), zap.Error(err))
	} else {
		logger.Info(
			fmt.Sprintf("Добавили в ES[%s] статью [%s] с id=[%s] от [%s]",
				ArticleIndex, a.Name, res.Id_, a.Publisher.Name))
	}

	return res.Result == result.Created
}

func (r *repository) IndexCategory(ctx context.Context, a *article.CategoryES) bool {
	h := fnv.New32a()
	if _, err := h.Write([]byte(a.Name)); err != nil {
		logger.Error(fmt.Sprintf("Не записали [%s] данные в [%s] error=[%v]", a.Name, CategoryIndex, err))
		return false
	}
	hash := h.Sum32()

	res, err := r.client.Index(CategoryIndex).
		Id(fmt.Sprintf("%d", hash)).
		Request(a).
		OpType(optype.Index).
		Do(ctx)

	if err != nil {
		logger.Error(fmt.Sprintf("Не записали [%s][%d] данные в [%s] error=[%v]", a.Name, hash, CategoryIndex, err))
	} else {
		logger.Info(fmt.Sprintf("Добавили в ES[%s] категорию [%s] с id=[%s]", CategoryIndex, a.Name, res.Id_))
	}

	return res.Result == result.Created
}

func (r *repository) IndexPeople(ctx context.Context, a *article.PersonES) bool {
	h := fnv.New32a()
	if _, err := h.Write([]byte(a.FullName)); err != nil {
		logger.Error(fmt.Sprintf("Не записали [%s] данные в [%s] error=[%v]", a.FullName, PeopleIndex, err))
		return false
	}
	hash := h.Sum32()

	res, err := r.client.Index(PeopleIndex).
		Id(fmt.Sprintf("%d", hash)).
		Request(a).
		OpType(optype.Index).
		Do(ctx)

	if err != nil {
		logger.Error(fmt.Sprintf("Не записали [%s][%d] данные в [%s] error=[%v]", a.FullName, hash, PeopleIndex, err))
	} else {
		logger.Info(fmt.Sprintf("Добавили в ES[%s] человека [%s] с id=[%s]", PeopleIndex, a.FullName, res.Id_))
	}

	return res.Result == result.Created
}

func (r *repository) FindArticles(ctx context.Context, f dto.GrandFilterRequest, size int) ([]*article.EsArticleDBO, int64, error) {
	span := trace.SpanFromContext(ctx)
	var must []types.Query

	// 1. Поисковые строки
	for i, filterString := range f.FilterStrings {
		q := &types.Query{Bool: &types.BoolQuery{}}
		s := strings.ToLower(filterString.Search)
		if filterString.IsExact {
			//q.CombinedFields = &types.CombinedFieldsQuery{
			//	Query:    s,
			//	Operator: &combinedfieldsoperator.And,
			//	Fields:   []string{"name", "description"},
			//}
			q.Bool.Should = []types.Query{
				{Match: map[string]types.MatchQuery{"name": {Query: s, Operator: &operator.And}}},
				{Match: map[string]types.MatchQuery{"description": {Query: s, Operator: &operator.And}}},
			}
		} else {
			// Разбиваю строку
			fuzzyMustName := &types.BoolQuery{Must: []types.Query{}}
			fuzzyMustDescription := &types.BoolQuery{Must: []types.Query{}}

			for _, word := range strings.Split(s, " ") {
				fuzzyMustName.Must = append(fuzzyMustName.Must,
					types.Query{Fuzzy: map[string]types.FuzzyQuery{"name": {Value: word}}})
				fuzzyMustDescription.Must = append(fuzzyMustDescription.Must,
					types.Query{Fuzzy: map[string]types.FuzzyQuery{"description": {Value: word}}})
			}

			q.Bool.Should = []types.Query{
				{Bool: fuzzyMustName},
				{Bool: fuzzyMustDescription},
				//{Fuzzy: map[string]types.FuzzyQuery{"name": {Value: s}}},
				//{Fuzzy: map[string]types.FuzzyQuery{"description": {Value: s}}},
			}
		}
		span.SetAttributes(attribute.String(
			fmt.Sprintf("Ищем строку №%d", i),
			fmt.Sprintf("Строка=[%s] Точный поиск=[%t]", s, filterString.IsExact)))
		must = append(must, *q)
	}

	// 2. Страна
	countryMust := types.Query{Bool: &types.BoolQuery{}}
	for i, country := range f.FilterCountry {
		q := types.Query{Match: map[string]types.MatchQuery{
			"address.country": {Query: country.Country}},
		}
		countryMust.Bool.Should = append(countryMust.Bool.Should, q)
		span.SetAttributes(attribute.String(
			fmt.Sprintf("Ищем страну №%d", i),
			fmt.Sprintf("Страна=[%s]", country)))
	}
	if len(f.FilterCountry) > 0 {
		must = append(must, countryMust)
	}

	// 3. Издание
	pubMust := types.Query{Bool: &types.BoolQuery{}}
	for i, p := range f.FilterPublishers {
		q := types.Query{Match: map[string]types.MatchQuery{
			"publisher.name": {Query: p.Name},
		}}
		pubMust.Bool.Should = append(pubMust.Bool.Should, q)
		span.SetAttributes(attribute.String(
			fmt.Sprintf("Ищем издание №%d", i),
			fmt.Sprintf("Издание=[%s]", p.Name)))
	}
	if len(f.FilterPublishers) > 0 {
		must = append(must, pubMust)
	}

	// 4. Люди
	peopleMust := types.Query{Bool: &types.BoolQuery{}}
	for i, person := range f.FilterPeople {
		q := types.Query{Match: map[string]types.MatchQuery{
			"people.fullName": {Query: person.FullName},
		}}
		peopleMust.Bool.Should = append(pubMust.Bool.Should, q)
		span.SetAttributes(attribute.String(
			fmt.Sprintf("Ищем человека №%d", i),
			fmt.Sprintf("Имя=[%s]", person.FullName)))
	}
	if len(f.FilterPeople) > 0 {
		must = append(must, peopleMust)
	}

	// 5. Категории
	categoriesMust := types.Query{Bool: &types.BoolQuery{}}
	for i, category := range f.FilterCategories {
		q := types.Query{Match: map[string]types.MatchQuery{
			"categories": {Query: category.Name},
		}}
		categoriesMust.Bool.Should = append(categoriesMust.Bool.Should, q)
		span.SetAttributes(attribute.String(
			fmt.Sprintf("Ищем категорию №%d", i),
			fmt.Sprintf("Категория=[%s]", category.Name)))
	}
	if len(f.FilterCategories) > 0 {
		must = append(must, categoriesMust)
	}

	// 6. Языки
	languagesMust := types.Query{Bool: &types.BoolQuery{}}
	for i, language := range f.FilterLanguages {
		q := types.Query{Match: map[string]types.MatchQuery{
			"language": {Query: language.Name},
		}}
		languagesMust.Bool.Should = append(languagesMust.Bool.Should, q)
		span.SetAttributes(attribute.String(
			fmt.Sprintf("Ищем язык №%d", i),
			fmt.Sprintf("Язык=[%s]", language.Name)))
	}
	if len(f.FilterLanguages) > 0 {
		must = append(must, languagesMust)
	}

	// 7. Время f.FilterTime

	req := &search.Request{
		Size: &size,
		Sort: types.Sort{
			map[string]types.FieldSort{
				"datePublished": {
					NumericType: lib.PointerFrom(fieldsortnumerictype.Date),
					Order:       lib.PointerFrom(sortorder.Desc),
				},
			},
		},
	}

	if len(must) > 0 {
		req.Query = &types.Query{
			Bool: &types.BoolQuery{Must: must},
		}
	}

	resp, err := r.client.Search().
		Index(ArticleIndex).
		Request(req).
		Do(ctx)
	if err != nil {
		return nil, 0, err
	}

	logger.Info(fmt.Sprintf("По запросу grandFilter нашли [%d]", resp.Hits.Total.Value))
	span.SetAttributes(attribute.Int64(
		fmt.Sprintf("Найдено"),
		resp.Hits.Total.Value))

	var p []*article.EsArticleDBO
	for _, hit := range resp.Hits.Hits {
		var res *article.EsArticleDBO
		if err := json.Unmarshal(hit.Source_, &res); err != nil {
			return nil, 0, err
		}
		p = append(p, res)
	}

	return p, resp.Hits.Total.Value, nil
}

func (r *repository) FindLanguages(ctx context.Context) ([]*article.LanguageES, error) {
	span := trace.SpanFromContext(ctx)
	req := &search.Request{
		Size: lib.PointerFrom(0),
		Aggregations: map[string]types.Aggregations{
			"langs": {
				Terms: &types.TermsAggregation{
					Field: lib.PointerFrom("language"),
				},
			},
		},
	}

	resp, err := r.client.Search().
		Index(ArticleIndex).
		Request(req).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Языков нашли [%d]", len(resp.Aggregations)))
	span.SetAttributes(attribute.Int(fmt.Sprintf("Найдено"), len(resp.Aggregations)))

	languages := resp.Aggregations["langs"].(*types.StringTermsAggregate)
	buckets := languages.Buckets.([]types.StringTermsBucket)
	fmt.Printf("%v", buckets)

	var p []*article.LanguageES
	for _, hit := range buckets {
		p = append(p, &article.LanguageES{Name: hit.Key.(string)})
	}

	return p, nil
}

func (r *repository) FindCategory(ctx context.Context, cat string) ([]*article.CategoryES, error) {
	span := trace.SpanFromContext(ctx)

	req := &search.Request{
		Size: lib.PointerFrom(20),
	}

	if len(cat) > 0 {
		span.SetAttributes(attribute.String("name:q", strings.ToLower(cat)))
		req.Query = &types.Query{Match: map[string]types.MatchQuery{
			"name": {Query: strings.ToLower(cat)},
		}}
	}

	resp, err := r.client.Search().
		Index(CategoryIndex).
		Request(req).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Нашли категорий [%d]", resp.Hits.Total.Value))
	span.SetAttributes(attribute.Int64(
		fmt.Sprintf("Найдено"),
		resp.Hits.Total.Value))

	var p []*article.CategoryES
	for _, hit := range resp.Hits.Hits {
		var res *article.CategoryES
		if err := json.Unmarshal(hit.Source_, &res); err != nil {
			return nil, err
		}
		p = append(p, res)
	}

	return p, nil
}

func (r *repository) FindPeople(ctx context.Context, name string) ([]*article.PersonES, error) {
	span := trace.SpanFromContext(ctx)

	req := &search.Request{
		Size: lib.PointerFrom(20),
	}

	if len(name) > 0 {
		span.SetAttributes(attribute.String("name:q", strings.ToLower(name)))
		req.Query = &types.Query{Match: map[string]types.MatchQuery{
			"fullName": {Query: strings.ToLower(name)},
		}}
	}

	resp, err := r.client.Search().
		Index(PeopleIndex).
		Request(req).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("Нашли людей [%d]", resp.Hits.Total.Value))
	span.SetAttributes(attribute.Int64(
		fmt.Sprintf("Найдено"),
		resp.Hits.Total.Value))

	var p []*article.PersonES
	for _, hit := range resp.Hits.Hits {
		var res *article.PersonES
		if err := json.Unmarshal(hit.Source_, &res); err != nil {
			return nil, err
		}
		p = append(p, res)
	}

	return p, nil
}
