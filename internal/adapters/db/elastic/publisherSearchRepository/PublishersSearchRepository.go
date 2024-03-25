package publishersSearchRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/result"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/tokenchar"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strings"
	"time"
)

const Index = "publisher"

var logger = logging.GetLogger().With(zap.String("prefix", "[ES]"))

type repository struct {
	client *elasticsearch.TypedClient
}

func New(client *elasticsearch.TypedClient) IRepository {
	return &repository{client}
}

type DTO struct {
	AddDate time.Time `json:"add_date"`
	Name    string    `json:"name"`
}

func (r *repository) Setup(ctx context.Context) {
	// 1. Создать индекс публициста
	if r.Exists(ctx) {
		logger.Info(fmt.Sprintf("Индекс %s уже существует, удаляем", Index))
		r.Delete(ctx)
	}

	if err := r.Create(ctx); err != nil {
		logger.Fatal("Проблема с индексом публициста", zap.Error(err))
	}

	// 2. Закинуть дефолтные значения
	publishers := []publisher.EsDBO{
		{Name: "The New York Times"},
		{Name: "The Guardian"},
		{Name: "Vedomosti"},
		{Name: "ООН"},
		{Name: "Hindustan Times"},
		{Name: "Rambler"},
		{Name: "lenta.ru"},
		{Name: "Wall Street Journal"},
		{Name: "France 24"},
		{Name: "CNN"},
		{Name: "Meduza"},
	}
	for _, p := range publishers {
		if ok := r.IndexPublisher(ctx, &p); !ok {
			logger.Fatal("Не записали данные в " + Index)
		}
	}

	// 3. Неточный поиск
	time.Sleep(2 * time.Second)
	if p, err := r.FindPublishersByNameViaES(ctx, "the new"); err != nil {
		logger.Fatal("Не нашли...", zap.Error(err))
	} else {
		for _, dbo := range p {
			logger.Info(fmt.Sprintf("Нашли %s с id=[%s] добавленный %s", dbo.Name, dbo.PublisherID, dbo.AddDate))
		}
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
	// Разбивает предложения по 2 буквы, включая пробелы
	// Для поиска названий
	MyTokenizer := types.NGramTokenizer{
		MinGram:    2,
		MaxGram:    30,
		TokenChars: []tokenchar.TokenChar{tokenchar.Letter, tokenchar.Digit, tokenchar.Whitespace},
		Type:       "ngram",
	}
	//logger.Info(types.NewLowercaseTokenFilter().Type) // -> "lowercase"
	res, err := r.client.Indices.Create(Index).
		Request(&create.Request{
			Settings: &types.IndexSettings{
				Analysis: &types.IndexSettingsAnalysis{
					Tokenizer: map[string]types.Tokenizer{
						"publisher_tokenizer": MyTokenizer,
					},
					Analyzer: map[string]types.Analyzer{
						"publisher_analyzer": types.CustomAnalyzer{
							Tokenizer: "publisher_tokenizer",
							Filter:    []string{"lowercase"},
						},
						"publisher_search_analyzer": types.CustomAnalyzer{
							Tokenizer: "keyword",
							Type:      "custom",
							Filter:    []string{"lowercase"},
						},
					},
				},
				MaxNgramDiff: lib.PointerFrom(30),
			},
			Mappings: &types.TypeMapping{
				Properties: map[string]types.Property{
					"name": &types.TextProperty{
						Analyzer:       lib.PointerFrom("publisher_analyzer"),
						SearchAnalyzer: lib.PointerFrom("publisher_search_analyzer"),
						Type:           "text",
						Index:          lib.PointerFrom(true),
					},
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

func (r *repository) IndexPublisher(ctx context.Context, p *publisher.EsDBO) bool {
	res, err := r.client.Index(Index).
		Request(DTO{AddDate: time.Now(), Name: p.Name}).
		Do(ctx)

	if err != nil {
		logger.Error("Не записали данные в "+Index, zap.Error(err))
	} else {
		p.PublisherID = res.Id_
		p.AddDate = time.Now()
		logger.Info(fmt.Sprintf("Добавили %s в ES[%s] с id=[%s]", p.Name, Index, res.Id_))
	}
	return res.Result == result.Created
}

func (r *repository) FindPublishersByNameViaES(ctx context.Context, name string) ([]*publisher.EsDBO, error) {
	var req *search.Request
	if name == "" {
		req = &search.Request{Size: lib.PointerFrom(5)}
	} else {
		req = &search.Request{
			Size: lib.PointerFrom(5),
			Query: &types.Query{
				//Match: map[string]types.MatchQuery{
				//	"name": {Query: strings.ToLower(name)},
				//},
				Fuzzy: map[string]types.FuzzyQuery{
					"name": {Value: strings.ToLower(name)},
				},
			},
		}
	}

	resp, err := r.client.Search().
		Index(Index).
		Request(req).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("По запросу [%s] нашли [%d]", name, resp.Hits.Total.Value))
	var p []*publisher.EsDBO
	for _, hit := range resp.Hits.Hits {
		var res DTO
		if err := json.Unmarshal(hit.Source_, &res); err != nil {
			return nil, err
		}
		p = append(p, &publisher.EsDBO{
			PublisherID: hit.Id_,
			AddDate:     res.AddDate,
			Name:        res.Name,
		})
	}

	return p, nil
}
