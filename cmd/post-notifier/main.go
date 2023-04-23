package main

import (
	"encoding/json"
	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/internal/pkg/contextlib"
	"github.com/antonpriyma/otus-highload/internal/pkg/middleware"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoapi"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoerrors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoutils"
	"github.com/antonpriyma/otus-highload/pkg/framework/service"
	"github.com/antonpriyma/otus-highload/pkg/utils"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type AppConfig struct {
	service.Config `mapstructure:",squash"`
}

func (a AppConfig) APIConfig() echoapi.Config {
	return echoapi.Config{
		ServeConfig: service.ServeConfig{
			GracefulWait: time.Second,
			StopWait:     time.Second,
		},
		Listen: ":8082",
	}
}

func main() {
	var cfg AppConfig
	cfg.Version = service.Version{
		Dist:    "local",
		Release: "local",
	}

	svc := echoapi.New(&cfg)
	conn, err := amqp.Dial("amqp://otus:otus@rabbitmq:5672/otus")
	utils.Must(svc.Logger, err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	utils.Must(svc.Logger, err, "Failed to open a channel")
	defer ch.Close()

	svc.API.Use(middleware.AuthMiddleware)
	svc.API.GET("/post/feed/posted", func(c echo.Context) error {
		userID, ok := contextlib.GetUserID(echoutils.MustGetContext(c))
		if !ok {
			return echoerrors.ValidationError(errors.New("user id not found"), "user id not found", echoerrors.ValidationErrorFields{})
		}

		upgrader := websocket.Upgrader{}
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer ws.Close()

		q, err := ch.QueueDeclare(
			"",    // name
			false, // durable
			true,  // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return echoerrors.InternalError(err)
		}

		err = ch.QueueBind(
			q.Name, // queue name
			string(userID),
			"post-notifications", // exchange
			false,
			nil,
		)
		if err != nil {
			return echoerrors.InternalError(err)
		}

		msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		if err != nil {
			return echoerrors.InternalError(err)
		}

		var Post models.Post
		for msg := range msgs {
			// Write
			err := json.Unmarshal(msg.Body, &Post)
			if err != nil {
				return echoerrors.InternalError(err)
			}

			err = ws.WriteJSON(Post)
		}

		return nil
	})

	svc.Run()
}
