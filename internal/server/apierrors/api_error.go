package apierrors

// APIError is the standard JSON error shape.
// Compatibility: response must include at least `{"error": "..."}`.
type APIError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"error"`
	Details any    `json:"details,omitempty"`
}

func (e APIError) Error() string {
	return e.Message
}
