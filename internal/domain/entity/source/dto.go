package source

import (
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
	"github.com/mskKote/prospero_backend/pkg/lib"
)

type AddSourceDTO struct {
	RssURL      string `json:"rss_url"`
	PublisherID string `json:"publisher_id"`
}

type DeleteSourceDTO struct {
	RssID string `json:"rss_id"`
}

type DTO struct {
	RssID       string `json:"rss_id"`
	RssURL      string `json:"rss_url"`
	PublisherID string `json:"publisher_id"`
}

func (dto *DTO) ToDomain() RSS {
	return RSS{
		RssID:     lib.StringToUUID(dto.RssID),
		RssURL:    dto.RssURL,
		Publisher: publisher.Publisher{PublisherID: lib.StringToUUID(dto.PublisherID)},
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
	}
}

func ToDTOs(r []*RSS) (d []*DTO) {
	for _, rss := range r {
		d = append(d, rss.ToDTO())
	}
	return d
}
