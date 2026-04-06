package pipeline

import (
	"encoding/json"
	"os"

	"olympics-planner/internal/domain"
)

// WriteSessionsJSON writes a JSON array of sessions in the format expected by
// repository.JSONSessionRepository and test fixtures.
func WriteSessionsJSON(path string, sessions []domain.Session) error {
	b, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
