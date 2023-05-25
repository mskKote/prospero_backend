package publisher

import "time"

// EsDBO TODO: сущность в ES
type EsDBO struct {
	PublisherID string    `json:"publisher_id"`
	AddDate     time.Time `json:"add_date"`
	Name        string    `json:"name"`
	//Country string `json:"country"`
	//City    string `json:"city"`
	//Point       elastic..Point     `json:"point"`
}
