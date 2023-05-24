package source

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
	"github.com/mskKote/prospero_backend/pkg/lib"
	"time"
)

type AddSourceAndPublisherDTO struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	RssUrl    string  `json:"rss_Url"`
}

type AddSourceDTO struct {
	RssURL      string `json:"rss_url"`
	PublisherID string `json:"publisher_id"`
}

type DeleteSourceDTO struct {
	RssID string `json:"rss_id"`
}

type DTO struct {
	RssID       string    `json:"rss_id"`
	RssURL      string    `json:"rss_url"`
	PublisherID string    `json:"publisher_id"`
	AddDate     time.Time `json:"add_date"`
}

func (dto *DTO) ToDomain() RSS {
	return RSS{
		RssID:     lib.StringToUUID(dto.RssID),
		RssURL:    dto.RssURL,
		Publisher: publisher.Publisher{PublisherID: lib.StringToUUID(dto.PublisherID)},
		AddDate: pgtype.Timestamp{
			Time:  dto.AddDate,
			Valid: true,
		},
	}
}

//func ToDomainMany(dtos []DTO) (r []RSS) {
//	for _, dto := range dtos {
//		r = append(r, dto.ToDomain())
//	}
//	return r
//}

func (r *RSS) ToDTO() *DTO {
	return &DTO{
		RssID:       lib.UuidToString(r.RssID),
		RssURL:      r.RssURL,
		PublisherID: lib.UuidToString(r.Publisher.PublisherID),
		AddDate:     r.AddDate.Time,
	}
}

func ToDTOs(r []*RSS) (d []*DTO) {
	for _, rss := range r {
		d = append(d, rss.ToDTO())
	}
	return d
}

type WithPublisher struct {
	Source    *DTO
	Publisher *publisher.DTO
}
