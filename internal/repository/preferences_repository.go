package repository

import "olympics-planner/internal/domain"

type PreferencesRepository interface {
    Get() (domain.Preferences, error)
}
