package mysql

import "github.com/antonpriyma/otus-highload/internal/app/models"

type Message struct {
	SenderUUID   string `db:"sender_uuid"`
	ReceiverUUID string `db:"receiver_uuid"`
	Text         string `db:"text"`
}

func convertModelToMessage(model models.Message) Message {
	return Message{
		SenderUUID:   string(model.From),
		ReceiverUUID: string(model.To),
		Text:         model.Text,
	}
}

func convertMessageToModel(message Message) models.Message {
	return models.Message{
		From: models.UserID(message.SenderUUID),
		To:   models.UserID(message.ReceiverUUID),
		Text: message.Text,
	}
}

func convertMessagesToModels(messages []Message) []models.Message {
	res := make([]models.Message, 0, len(messages))
	for _, message := range messages {
		res = append(res, convertMessageToModel(message))
	}

	return res
}
