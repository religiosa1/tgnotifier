package models

type ResponsePayload struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
