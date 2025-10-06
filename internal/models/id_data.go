package models

import "time"

// IDData represents extracted ID information
type IDData struct {
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	BirthDate string    `json:"birth_date"`
	IDNP      string    `json:"idnp"` // Personal identification number
	RawText   string    `json:"raw_text"`
	Timestamp time.Time `json:"timestamp"`
}

// ScanRequest represents incoming image scan request
type ScanRequest struct {
	RequestID string `json:"request_id"`
	ImageData []byte `json:"image_data"` // Base64 encoded or raw bytes
	ImagePath string `json:"image_path"` // Alternative: path to image file
}

// ScanResponse represents the response sent to Kafka
type ScanResponse struct {
	RequestID string  `json:"request_id"`
	Success   bool    `json:"success"`
	Data      *IDData `json:"data,omitempty"`
	Error     string  `json:"error,omitempty"`
}
