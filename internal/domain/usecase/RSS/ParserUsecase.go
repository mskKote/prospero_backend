package RSS

import (
	"context"
	"github.com/go-co-op/gocron"
	"github.com/mskKote/prospero_backend/internal/domain/service/articleService"
	"github.com/mskKote/prospero_backend/internal/domain/service/sourcesService"
	"github.com/mskKote/prospero_backend/pkg/config"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
	"time"
)

var (
	logger = logging.GetLogger()
	cfg    = config.GetConfig()
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

	if _, err := s.Cron(cfg.CronSourcesRSS).Do(u.ParseJob); err != nil {
		logger.Fatal("Не стартовали CRON RSS", zap.Error(err))
	}

	// При миграции всё удаляется, достаю данные
	if cfg.MigrateElastic {
		logger.Info("[MIGRATION] Закинуть данные в ELASTIC сразу")
		u.ParseJob()
	}

	s.StartAsync()
}

func (u *usecase) ParseJob() {
	ctx := context.Background()
	if err := u.articles.ParseAllOnce(ctx); err != nil {
		return
	}
}
