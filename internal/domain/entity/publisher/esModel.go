package publisher

import "time"

// EsDBO сущность публициста в ES
type EsDBO struct {
	PublisherID string    `json:"publisher_id"`
	AddDate     time.Time `json:"add_date"`
	Name        string    `json:"name"`
}
