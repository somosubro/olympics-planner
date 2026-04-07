package http

// ErrorBody is the nested object for 400 responses (api-spec §14).
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type errorEnvelope struct {
	Error ErrorBody `json:"error"`
}
