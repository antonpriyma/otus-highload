package notifer

import (
	"context"
	"encoding/json"
	"github.com/antonpriyma/otus-highload/internal/app/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Notifer struct {
	ch *amqp.Channel
}

func NewNotifer(ch *amqp.Channel) (Notifer, error) {
	err := ch.ExchangeDeclare(
		"post-notifications", // name
		"direct",             // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)

	if err != nil {
		return Notifer{}, err
	}
	return Notifer{
		ch: ch,
	}, nil
}

func (n Notifer) Notify(ctx context.Context, post models.Post, userID models.UserID) error {
	body, err := json.Marshal(post)
	if err != nil {
		return err
	}

	err = n.ch.PublishWithContext(
		ctx,
		"post-notifications", // exchange
		string(userID),       // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return err
	}

	return nil
}
