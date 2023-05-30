package articlesSearchRepository

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/domain/entity/article"
)

type IRepository interface {
	Setup(ctx context.Context)
	Exists(ctx context.Context) bool
	Delete(ctx context.Context)
	Create(ctx context.Context) error
	FindArticles(ctx context.Context, f dto.GrandFilterRequest) ([]*article.EsArticleDBO, error)
	IndexArticle(ctx context.Context, a *article.EsArticleDBO) bool
}
