package domain

// ValidationResult matches docs/data-contract.md §13 (validation response shape).
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors"`
}

// ValidationError matches docs/data-contract.md §13.3.
type ValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}
