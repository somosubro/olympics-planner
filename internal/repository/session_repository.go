package repository

import "olympics-planner/internal/domain"

type SessionRepository interface {
	GetAll() ([]domain.Session, error)
}
