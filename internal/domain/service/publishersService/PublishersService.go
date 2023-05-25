package publishersService

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	publishersSearchRepository "github.com/mskKote/prospero_backend/internal/adapters/db/elastic/publisherSearchRepository"
	"github.com/mskKote/prospero_backend/internal/adapters/db/postgres/publishersRepository"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"time"
)

type service struct {
	postgres publishersRepository.IRepository
	elastic  publishersSearchRepository.IRepository
}

func (s *service) Create(ctx context.Context, addDTO *publisher.AddPublisherDTO) (*publisher.DTO, error) {
	dto := publisher.DTO{
		PublisherID: "",
		AddDate:     time.Now(),
		Name:        addDTO.Name,
		Country:     addDTO.Country,
		City:        addDTO.City,
		Longitude:   addDTO.Longitude,
		Latitude:    addDTO.Latitude,
	}
	data, err := s.postgres.Create(ctx, dto.ToDomain())
	return data.ToDTO(), err
}

func (s *service) FindAll(ctx context.Context) ([]*publisher.DTO, error) {
	p, err := s.postgres.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return publisher.PgDBOsToDTOs(p), nil
}

func (s *service) FindPublishersByName(ctx context.Context, name string) ([]*publisher.DTO, error) {
	p, err := s.postgres.FindPublishersByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return publisher.PgDBOsToDTOs(p), nil
}

func (s *service) FindPublishersByIDs(ctx context.Context, ids []string) ([]*publisher.DTO, error) {
	var uuids []pgtype.UUID
	for _, id := range ids {
		uuids = append(uuids, lib.StringToUUID(id))
	}

	p, err := s.postgres.FindPublishersByIDs(ctx, uuids)
	if err != nil {
		return nil, err
	}

	return publisher.PgDBOsToDTOs(p), nil
}

func (s *service) FindPublishersByNameViaES(ctx context.Context, name string) ([]*publisher.DTO, error) {
	p, err := s.elastic.FindPublishersByNameViaES(ctx, name)
	if err != nil {
		return nil, err
	}
	return publisher.EsDBOsToDTOs(p), nil
}

func (s *service) Update(ctx context.Context, dto *publisher.DTO) error {
	return s.postgres.Update(ctx, dto.ToDomain())
}

func (s *service) Delete(ctx context.Context, dto *publisher.DeletePublisherDTO) error {
	return s.postgres.Delete(ctx, dto.PublisherID)
}

func New(
	postgres publishersRepository.IRepository,
	elastic publishersSearchRepository.IRepository) IPublishersService {
	return &service{postgres, elastic}
}
