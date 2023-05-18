package sourcesRepository

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/domain/entity/source"
)

type IRepository interface {
	Create(ctx context.Context, source *source.RSS) (*source.RSS, error)
	FindAll(ctx context.Context) ([]*source.RSS, error)
	FindByPublisherName(ctx context.Context, name string) ([]*source.RSS, error)
	Update(ctx context.Context, source source.RSS) error
	Delete(ctx context.Context, id string) error
}
