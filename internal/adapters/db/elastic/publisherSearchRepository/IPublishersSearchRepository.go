package publishersSearchRepository

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
)

type IRepository interface {
	FindPublishersByNameViaES(ctx context.Context, name string) ([]*publisher.EsDBO, error)
	IndexPublisher(ctx context.Context, p *publisher.EsDBO) bool
}
