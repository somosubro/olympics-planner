package repository

import (
	"encoding/json"
	"os"

	"olympics-planner/internal/domain"
)

type JSONPreferencesRepository struct {
	Path string
}

func (r JSONPreferencesRepository) Get() (domain.Preferences, error) {
	data, err := os.ReadFile(r.Path)
	if err != nil {
		return domain.Preferences{}, err
	}

	var preferences domain.Preferences
	if err := json.Unmarshal(data, &preferences); err != nil {
		return domain.Preferences{}, err
	}

	return preferences, nil
}
