package usecase

import (
	"context"
	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

type usecase struct {
	logger  log.Logger
	dialogs models.DialogRepository
}

func (u usecase) SendMessage(ctx context.Context, message models.Message) error {
	err := u.dialogs.SendMessage(ctx, message)
	if err != nil {
		return errors.Wrap(err, "failed to save message to repository")
	}

	return nil
}

func (u usecase) GetDialog(ctx context.Context, userID models.UserID, friendID models.UserID) ([]models.Message, error) {
	dialog, err := u.dialogs.GetDialog(ctx, userID, friendID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get dialog from repository")
	}

	return dialog, nil
}

func NewUsecase(dialogs models.DialogRepository, logger log.Logger) models.DialogUsecase {
	return usecase{
		logger:  logger,
		dialogs: dialogs,
	}
}
