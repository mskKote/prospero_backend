package publisher

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"time"
)

type AddPublisherDTO struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type DeletePublisherDTO struct {
	PublisherID string `json:"publisher_id"`
}

type DTO struct {
	PublisherID string    `json:"publisher_id"`
	AddDate     time.Time `json:"add_date"`
	Name        string    `json:"name"`
	Country     string    `json:"country"`
	City        string    `json:"city"`
	Longitude   float64   `json:"longitude"`
	Latitude    float64   `json:"latitude"`
}

func (p *Publisher) ToDTO() *DTO {
	return &DTO{
		PublisherID: lib.UuidToString(p.PublisherID),
		AddDate:     p.AddDate.Time,
		Name:        p.Name,
		Country:     p.Country,
		City:        p.City,
		Longitude:   p.Point.P.X,
		Latitude:    p.Point.P.Y,
	}
}

func (dto *DTO) ToDomain() *Publisher {

	return &Publisher{
		PublisherID: lib.StringToUUID(dto.PublisherID),
		AddDate: pgtype.Timestamp{
			Time:  dto.AddDate,
			Valid: true,
		},
		Name:    dto.Name,
		Country: dto.Country,
		City:    dto.City,
		Point: pgtype.Point{
			P: pgtype.Vec2{
				X: dto.Longitude,
				Y: dto.Latitude,
			},
			Valid: true,
		},
	}
}

func ToDTOs(p []*Publisher) (d []*DTO) {
	for _, publisher := range p {
		d = append(d, publisher.ToDTO())
	}
	return d
}
