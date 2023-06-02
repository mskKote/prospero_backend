package sourcesService

import (
	"context"
	"github.com/mskKote/prospero_backend/internal/adapters/db/postgres/sourcesRepository"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
	"github.com/mskKote/prospero_backend/internal/domain/entity/source"
	"github.com/mskKote/prospero_backend/pkg/lib"
)

type service struct {
	sources sourcesRepository.IRepository
}

func (s *service) Count(ctx context.Context) (int64, error) {
	return s.sources.Count(ctx)
}

func (s *service) FindAll(ctx context.Context, page, pageSize int) ([]*source.DTO, error) {

	sourcesRSS, err := s.sources.FindAll(ctx, page*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	return source.ToDTOs(sourcesRSS), nil
}

func (s *service) FindAllWithPublisher(ctx context.Context, page, pageSize int) ([]*source.RSS, error) {
	return s.sources.FindAllWithPublishers(ctx, page*pageSize, pageSize)
}

func (s *service) FindByPublisherName(ctx context.Context, name string, page, pageSize int) (sec []*source.DTO, err error) {
	src, err := s.sources.FindByPublisherName(ctx, name, page*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	return source.ToDTOs(src), nil
}

func (s *service) Update(ctx context.Context, dto *source.DTO) (*source.DTO, error) {
	err := s.sources.Update(ctx, dto.ToDomain())
	return dto, err
}

func (s *service) Delete(ctx context.Context, dto source.DeleteSourceDTO) error {
	return s.sources.Delete(ctx, dto.RssID)
}

func (s *service) AddSource(ctx context.Context, dto source.AddSourceDTO) (*source.DTO, error) {
	id := lib.StringToUUID(dto.PublisherID)
	p := publisher.PgDBO{PublisherID: id}

	saved, err := s.sources.Create(ctx, &source.RSS{
		RssURL:    dto.RssURL,
		Publisher: p,
	})
	return saved.ToDTO(), err
}

func New(sources sourcesRepository.IRepository) ISourceService {
	return &service{sources}
}
