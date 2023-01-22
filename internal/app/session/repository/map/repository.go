package map_repository

import (
	"context"
	"sync"

	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/google/uuid"
)

type MapRepository struct {
	sessions sync.Map
	logger   log.Logger
}

func NewSessionRepository(logger log.Logger) models.SessionRepository {
	return &MapRepository{
		logger: logger,
	}
}

func (m *MapRepository) CreateSession(_ context.Context, userID models.UserID) (models.SessionToken, error) {
	token := uuid.New().String()
	m.sessions.Store(userID, token)
	return models.SessionToken(token), nil
}
