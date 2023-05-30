package adminService

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/domain/entity/admin"
)

type IAdminService interface {
	Create(ctx context.Context, dto *admin.DTO) error
	Login(ctx context.Context, dto *admin.DTO) (*admin.Admin, bool)
}
