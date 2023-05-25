package publishersRepository

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
)

type IRepository interface {
	Create(ctx context.Context, source *publisher.PgDBO) (*publisher.PgDBO, error)
	FindAll(ctx context.Context) (u []*publisher.PgDBO, err error)
	FindPublishersByName(ctx context.Context, name string) ([]*publisher.PgDBO, error)
	FindPublishersByIDs(ctx context.Context, ids []pgtype.UUID) ([]*publisher.PgDBO, error)
	Update(ctx context.Context, publisher *publisher.PgDBO) error
	Delete(ctx context.Context, id string) error
}
