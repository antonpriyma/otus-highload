package grpc

import (
	"context"
	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/dialogs/github.com/antonpriyma/otus-highload/pkg/dialogs"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

type dialogDelivery struct {
	logger  log.Logger
	dialogs models.DialogUsecase
	dialogs.UnimplementedDialogsServer
}

func (d dialogDelivery) SendMessage(ctx context.Context, request *dialogs.SendMessageRequest) (*dialogs.SendMessageResponse, error) {
	modelMessage := models.Message{
		From: models.UserID(request.Message.From),
		To:   models.UserID(request.Message.To),
		Text: request.Message.Text,
	}

	err := d.dialogs.SendMessage(ctx, modelMessage)
	if err != nil {
		return nil, err
	}

	return &dialogs.SendMessageResponse{}, nil
}

func (d dialogDelivery) GetMessages(ctx context.Context, request *dialogs.GetMessagesRequest) (*dialogs.GetMessagesResponse, error) {
	modelMessages, err := d.dialogs.GetDialog(ctx, models.UserID(request.User), models.UserID(request.From))
	if err != nil {
		return nil, err
	}

	var messages []*dialogs.Message
	for _, modelMessage := range modelMessages {
		messages = append(messages, &dialogs.Message{
			From: string(modelMessage.From),
			To:   string(modelMessage.To),
			Text: modelMessage.Text,
		})
	}

	return &dialogs.GetMessagesResponse{
		Messages: messages,
	}, nil
}

func NewDelivery(dialogs models.DialogUsecase, logger log.Logger) dialogs.DialogsServer {
	return dialogDelivery{
		logger:  logger,
		dialogs: dialogs,
	}
}
