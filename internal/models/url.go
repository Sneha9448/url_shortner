package models

import "time"

type ShortenRequest struct {
	URL string `json:"url"`
	TTL int    `json:"ttl,omitempty"` // TTL in seconds
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// URLDocument represents the MongoDB document
type URLDocument struct {
	ID          string     `bson:"_id,omitempty"`
	ShortCode   string     `bson:"short_code"`
	OriginalURL string     `bson:"original_url"`
	Clicks      int        `bson:"clicks"`
	CreatedAt   time.Time  `bson:"created_at"`
	ExpiresAt   *time.Time `bson:"expires_at,omitempty"`
}
