package elastic

import (
	"context"
	"fmt"
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
	conStr := fmt.Sprintf("http://%s:%s", cfg.Elastic.Host, cfg.Elastic.Port)
	logger.InfoContext(ctx, conStr)

	cfg := elasticsearch.Config{
		Addresses: []string{
			conStr,
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
