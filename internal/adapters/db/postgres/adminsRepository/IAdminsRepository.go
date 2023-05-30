package adminsRepository

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/domain/entity/admin"
)

type IRepository interface {
	Create(ctx context.Context, a *admin.Admin) error
	FindAdminByName(ctx context.Context, name string) (*admin.Admin, error)
}
