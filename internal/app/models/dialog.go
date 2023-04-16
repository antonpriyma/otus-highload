package models

import "context"

type Message struct {
	From UserID
	To   UserID
	Text string
}

type DialogDelivery interface {
	SendMessage(ctx context.Context, message Message) error
	GetDialog(ctx context.Context, userID UserID, friendID UserID) ([]Message, error)
}

type DialogUsecase interface {
	SendMessage(ctx context.Context, message Message) error
	GetDialog(ctx context.Context, userID UserID, friendID UserID) ([]Message, error)
}

type DialogRepository interface {
	SendMessage(ctx context.Context, message Message) error
	GetDialog(ctx context.Context, userID UserID, friendID UserID) ([]Message, error)
}
