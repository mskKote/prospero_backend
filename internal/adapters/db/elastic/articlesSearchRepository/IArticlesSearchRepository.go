package articlesSearchRepository

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/domain/entity/article"
)

type IRepository interface {
	Setup(ctx context.Context)
	Exists(ctx context.Context, index string) bool
	Delete(ctx context.Context, index string)
	Create(ctx context.Context) error
	IndexArticle(ctx context.Context, a *article.EsArticleDBO) bool
	IndexCategory(ctx context.Context, a *article.CategoryES) bool
	IndexPeople(ctx context.Context, a *article.PersonES) bool
	FindArticles(ctx context.Context, f dto.GrandFilterRequest, size int) ([]*article.EsArticleDBO, int64, error)
	FindLanguages(ctx context.Context) ([]*article.LanguageES, error)
	FindCategory(ctx context.Context, cat string) ([]*article.CategoryES, error)
	FindPeople(ctx context.Context, name string) ([]*article.PersonES, error)
}
