package adminka

import "github.com/mskKote/prospero_backend/internal/controller/http/v1/routes"

type IAdminkaUsecase interface {
	routes.ISourcesUsecase
	routes.IPublishersUsecase
}
