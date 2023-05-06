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
}

func (d dialogDelivery) SendMessage(ctx context.Context, request *dialogs.SendMessageRequest) (*dialogs.SendMessageResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (d dialogDelivery) GetMessages(ctx context.Context, request *dialogs.GetMessagesRequest) (*dialogs.GetMessagesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (d dialogDelivery) mustEmbedUnimplementedDialogsServer() {
	//TODO implement me
	panic("implement me")
}

func NewDelivery(dialogs models.DialogUsecase, logger log.Logger) dialogs.DialogsServer {
	return dialogDelivery{
		logger:  logger,
		dialogs: dialogs,
	}
}
