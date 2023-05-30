package sourcesService

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/domain/entity/source"
)

type ISourceService interface {
	AddSource(ctx context.Context, dto source.AddSourceDTO) (*source.DTO, error)
	FindAll(ctx context.Context, page, pageSize int) ([]*source.DTO, error)
	FindAllWithPublisher(ctx context.Context, page, pageSize int) ([]*source.RSS, error)
	FindByPublisherName(ctx context.Context, name string, page, pageSize int) ([]*source.DTO, error)
	Update(ctx context.Context, source *source.DTO) (*source.DTO, error)
	Delete(ctx context.Context, dto source.DeleteSourceDTO) error
	Count(ctx context.Context) (int64, error)
}
