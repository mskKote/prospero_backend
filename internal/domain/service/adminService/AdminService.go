package adminService

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/adapters/db/postgres/adminsRepository"
	"github.com/mskKote/prospero_backend/internal/domain/entity/admin"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"github.com/mskKote/prospero_backend/pkg/logging"
	"go.uber.org/zap"
)

var logger = logging.GetLogger()

type service struct {
	repository adminsRepository.IRepository
}

func New(adminRepo adminsRepository.IRepository) IAdminService {
	return &service{adminRepo}
}

func (s service) Create(ctx context.Context, dto *admin.DTO) error {
	if hash, err := lib.HashPassword(dto.Password); err != nil {
		logger.Error("Не захэшировали пароль админа", zap.Error(err))
		return nil
	} else {
		dto.Password = hash
	}

	err := s.repository.Create(ctx, dto.ToDomain())
	if err != nil {
		logger.Error("Не создали админа", zap.Error(err))
		return err
	}
	return nil
}

func (s service) Login(ctx context.Context, dto *admin.DTO) (*admin.Admin, bool) {
	a, err := s.repository.FindAdminByName(ctx, dto.Name)
	if err != nil {
		logger.Info("Админа {" + dto.Name + "} не существует")
		return nil, false
	}
	if lib.CheckPasswordHash(dto.Password, a.Password) {
		return a, true
	} else {
		logger.Info("Неправильный пароль {" + dto.Password + "} для {" + dto.Name + "}")
		return nil, false
	}
}
