package sourcesService

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/domain/entity/source"
)

type ISourceService interface {
	AddSource(ctx context.Context, dto source.AddSourceDTO) (*source.DTO, error)
	FindAll(ctx context.Context) ([]*source.DTO, error)
	FindByPublisherName(ctx context.Context, name string) ([]*source.DTO, error)
	Update(ctx context.Context, source *source.DTO) (*source.DTO, error)
	Delete(ctx context.Context, dto source.DeleteSourceDTO) error
}
