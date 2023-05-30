package articleService

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/adapters/db/elastic/articlesSearchRepository"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/domain/entity/article"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"github.com/mskKote/prospero_backend/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

var logger = logging.GetLogger()

type service struct {
	elastic articlesSearchRepository.IRepository
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
	elastic articlesSearchRepository.IRepository) IArticleService {
	return &service{elastic}
}
