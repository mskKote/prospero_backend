package articleService

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/controller/http/v1/dto"
	"github.com/mskKote/prospero_backend/internal/domain/entity/article"
)

type IArticleService interface {
	AddArticle(ctx context.Context, dto *article.EsArticleDBO) bool
	FindWithGrandFilter(ctx context.Context, p dto.GrandFilterRequest) ([]*article.EsArticleDBO, error)
}
