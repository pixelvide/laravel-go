package queue

import "encoding/json"

// LaravelJob represents the standard JSON structure of a Laravel queue job
type LaravelJob struct {
	UUID          string          `json:"uuid"`
	DisplayName   string          `json:"displayName"`
	Job           string          `json:"job"`
	MaxTries      *int            `json:"maxTries"`
	MaxExceptions *int            `json:"maxExceptions"`
	Backoff       *int            `json:"backoff"`
	Timeout       *int            `json:"timeout"`
	Data          json.RawMessage `json:"data"`
	Attempts      int             `json:"attempts"` // Laravel often stores attempts internally or in payload
}
