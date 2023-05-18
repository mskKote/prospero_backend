package publisher

import "github.com/jackc/pgx/v5/pgtype"

type Publisher struct {
	PublisherID pgtype.UUID      `json:"publisher_id"`
	AddDate     pgtype.Timestamp `json:"add_date"`
	Name        string           `json:"name"`
	Country     string           `json:"country"`
	City        string           `json:"city"`
	Point       pgtype.Point     `json:"point"`
}
