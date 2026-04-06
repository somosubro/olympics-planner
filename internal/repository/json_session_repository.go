package repository

import (
    "encoding/json"
    "os"

    "olympics-planner/internal/domain"
)

type JSONSessionRepository struct {
    Path string
}

func (r JSONSessionRepository) GetAll() ([]domain.Session, error) {
    data, err := os.ReadFile(r.Path)
    if err != nil {
        return nil, err
    }

    var sessions []domain.Session
    if err := json.Unmarshal(data, &sessions); err != nil {
        return nil, err
    }

    return sessions, nil
}
