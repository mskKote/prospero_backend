package publishersSearchRepository

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
)

type IRepository interface {
	Setup(ctx context.Context)
	Exists(ctx context.Context) bool
	Delete(ctx context.Context)
	Create(ctx context.Context) error
	FindPublishersByNameViaES(ctx context.Context, name string) ([]*publisher.EsDBO, error)
	IndexPublisher(ctx context.Context, p *publisher.EsDBO) bool
}
