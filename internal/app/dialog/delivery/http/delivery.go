package http

import (
	"context"
	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

type dialogDelivery struct {
	logger  log.Logger
	dialogs models.DialogUsecase
}

func (d dialogDelivery) SendMessage(ctx context.Context, message models.Message) error {
	err := d.dialogs.SendMessage(ctx, message)
	if err != nil {
		return err
	}

	return nil
}

func (d dialogDelivery) GetDialog(ctx context.Context, userID models.UserID, friendID models.UserID) ([]models.Message, error) {
	dialog, err := d.dialogs.GetDialog(ctx, userID, friendID)
	if err != nil {
		return nil, err
	}

	return dialog, nil
}

func NewDialogDelivery(dialogs models.DialogUsecase, logger log.Logger) models.DialogDelivery {
	return dialogDelivery{
		logger:  logger,
		dialogs: dialogs,
	}
}
