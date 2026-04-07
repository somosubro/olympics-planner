package repository

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"olympics-planner/internal/domain"
)

// ErrSessionsFileNotJSONArray is returned when the file root is not a JSON array
// (for example an object wrapper { "sessions": [...] }).
var ErrSessionsFileNotJSONArray = errors.New("sessions file must be a JSON array at root, not an object wrapper")

type JSONSessionRepository struct {
	Path string
}

func (r JSONSessionRepository) GetAll() ([]domain.Session, error) {
	data, err := os.ReadFile(r.Path)
	if err != nil {
		return nil, err
	}
	return unmarshalSessionsJSONArray(data)
}

func unmarshalSessionsJSONArray(data []byte) ([]domain.Session, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("sessions file is empty")
	}
	if data[0] != '[' {
		return nil, ErrSessionsFileNotJSONArray
	}
	var sessions []domain.Session
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}
