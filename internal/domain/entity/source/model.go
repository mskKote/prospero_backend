package source

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mskKote/prospero_backend/internal/domain/entity/publisher"
)

type RSS struct {
	RssID     pgtype.UUID         `json:"rss_id"`
	RssURL    string              `json:"rss_url"`
	Publisher publisher.Publisher `json:"publisher_id"`
	AddDate   pgtype.Timestamp    `json:"add_date"`
}
