package articleService

import (
	"context"
	"github.com/mmcdole/gofeed"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/domain/entity/article"
)

type IArticleService interface {
	FindWithGrandFilter(ctx context.Context, p dto.GrandFilterRequest, size int) ([]*article.EsArticleDBO, int64, error)

	// ParseAllOnce проходит по всем источникам
	ParseAllOnce(ctx context.Context) error

	// ParseRSS достаёт контент по источнику
	ParseRSS(src string) *gofeed.Feed

	FindAllLanguages(ctx context.Context) ([]*article.LanguageES, error)

	FindCategory(ctx context.Context, cat string) ([]*article.CategoryES, error)

	FindPeople(ctx context.Context, name string) ([]*article.PersonES, error)
}
