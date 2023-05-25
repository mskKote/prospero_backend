package publishersService

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
)

type IPublishersService interface {
	Create(ctx context.Context, dto *publisher.AddPublisherDTO) (*publisher.DTO, error)
	FindAll(ctx context.Context) ([]*publisher.DTO, error)
	FindPublishersByName(ctx context.Context, name string) ([]*publisher.DTO, error)
	FindPublishersByNameViaES(ctx context.Context, name string) ([]*publisher.DTO, error)
	FindPublishersByIDs(ctx context.Context, ids []string) ([]*publisher.DTO, error)
	Update(ctx context.Context, dto *publisher.DTO) error
	Delete(ctx context.Context, dto *publisher.DeletePublisherDTO) error
}
