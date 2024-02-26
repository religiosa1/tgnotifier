package models

type ResponsePayload struct {
	RequestId string `json:"requestId"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}
