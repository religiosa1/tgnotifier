package models

type ResponsePayload struct {
	RequestId    string `json:"requestId"`
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error,omitempty"`
}

func (e *ResponsePayload) Error() string {
	return e.ErrorMessage
}
