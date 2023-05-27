package articleService

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/adapters/db/elastic/articlesSearchRepository"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/domain/entity/article"
)

type service struct {
	elastic articlesSearchRepository.IRepository
}

func (s *service) FindWithGrandFilter(ctx context.Context, p dto.GrandFilterRequest) ([]*article.EsArticleDBO, error) {
	return s.elastic.FindArticles(ctx, p)
}

func New(
	elastic articlesSearchRepository.IRepository) IArticleService {
	return &service{elastic}
}
