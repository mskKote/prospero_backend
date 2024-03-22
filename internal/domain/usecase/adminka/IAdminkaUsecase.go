package adminka

import "github.com/mskKote/prospero_backend/internal/controller/http/v1/routes"

type IAdminkaUseCase interface {
	routes.ISourcesUseCase
	routes.IPublishersUseCase
	routes.IServiceUseCase
}
