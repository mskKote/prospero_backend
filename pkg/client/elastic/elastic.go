package elastic

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
	"log"
)

var (
	logger = logging.GetLogger()
	cfg    = config.GetConfig()
)

func NewClient(ctx context.Context) (es *elasticsearch.TypedClient, err error) {
	logger.InfoContext(ctx, cfg.Elastic.ConStr)

	cfg := elasticsearch.Config{
		Addresses: []string{
			cfg.Elastic.ConStr,
		},
	}

	if es, err = elasticsearch.NewTypedClient(cfg); err != nil {
		logger.FatalContext(ctx, "Не подключились к ES", zap.Error(err))
	}
	logger.InfoContext(ctx, elasticsearch.Version)
	res, err := es.Info().Do(ctx)
	if err != nil {
		logger.FatalContext(ctx, "Не получили ответ от ES", zap.Error(err))
	}
	log.Println(res)

	return es, nil
}
