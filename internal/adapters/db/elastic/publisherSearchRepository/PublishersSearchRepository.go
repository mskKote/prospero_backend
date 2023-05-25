package publishersSearchRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
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

func (r *repository) IndexPublisher(ctx context.Context, p *publisher.EsDBO) bool {
	res, err := r.client.Index(Index).
		Request(DTO{AddDate: time.Now(), Name: p.Name}).
		Do(ctx)

	if err != nil {
		logger.Error("Не записали данные в "+Index, zap.Error(err))
	} else {
		p.PublisherID = res.Id_
		p.AddDate = time.Now()
		logger.Info(fmt.Sprintf("Добавили %s в ES[%s] с id=%s", p.Name, Index, res.Id_))
	}
	return err == nil
}

func (r *repository) FindPublishersByNameViaES(ctx context.Context, name string) ([]*publisher.EsDBO, error) {
	resp, err := r.client.Search().
		Index(Index).
		Request(&search.Request{
			Query: &types.Query{
				//Match: map[string]types.MatchQuery{
				//	"name": {Query: strings.ToLower(name)},
				//},
				Fuzzy: map[string]types.FuzzyQuery{
					"name": {Value: strings.ToLower(name)},
				},
			},
		}).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(fmt.Sprintf("По запросу %s нашли %d", name, resp.Hits.Total.Value))
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
