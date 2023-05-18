package publishersService

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mskKote/prospero_backend/internal/adapters/db/postgres/publishersRepository"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"time"
)

type service struct {
	repository publishersRepository.IRepository
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
	data, err := s.repository.Create(ctx, dto.ToDomain())
	return data.ToDTO(), err
}

func (s *service) FindAll(ctx context.Context) ([]*publisher.DTO, error) {
	p, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return publisher.ToDTOs(p), nil
}

func (s *service) FindPublishersByName(ctx context.Context, name string) ([]*publisher.DTO, error) {
	p, err := s.repository.FindPublishersByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return publisher.ToDTOs(p), nil
}

func (s *service) FindPublishersByIDs(ctx context.Context, ids []string) ([]*publisher.DTO, error) {
	var uuids []pgtype.UUID
	for _, id := range ids {
		uuids = append(uuids, lib.StringToUUID(id))
	}

	p, err := s.repository.FindPublishersByIDs(ctx, uuids)
	if err != nil {
		return nil, err
	}

	return publisher.ToDTOs(p), nil
}

func (s *service) Update(ctx context.Context, dto *publisher.DTO) error {
	return s.repository.Update(ctx, dto.ToDomain())
}

func (s *service) Delete(ctx context.Context, dto *publisher.DeletePublisherDTO) error {
	return s.repository.Delete(ctx, dto.PublisherID)
}

func New(publishersRepo publishersRepository.IRepository) IPublishersService {
	return &service{publishersRepo}
}
