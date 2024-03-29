package main

import (
	"context"
	dialog_repo "github.com/antonpriyma/otus-highload/internal/app/dialog/repository/mysql"
	"github.com/antonpriyma/otus-highload/internal/app/models"
	post_delivery "github.com/antonpriyma/otus-highload/internal/app/post/delivery/http"
	"github.com/antonpriyma/otus-highload/internal/app/post/notifer"
	post_repo "github.com/antonpriyma/otus-highload/internal/app/post/repository/mysql"
	post_usecase "github.com/antonpriyma/otus-highload/internal/app/post/usecase"
	map_repository "github.com/antonpriyma/otus-highload/internal/app/session/repository/map"
	user_delivery "github.com/antonpriyma/otus-highload/internal/app/user/delivery/http"
	user_repo "github.com/antonpriyma/otus-highload/internal/app/user/repository/mysql"
	"github.com/antonpriyma/otus-highload/internal/app/user/usecase"
	"github.com/antonpriyma/otus-highload/internal/pkg/contextlib"
	"github.com/antonpriyma/otus-highload/internal/pkg/middleware"
	"github.com/antonpriyma/otus-highload/pkg/context/reqid"
	"github.com/antonpriyma/otus-highload/pkg/dialogs/github.com/antonpriyma/otus-highload/pkg/dialogs"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoapi"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoerrors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoutils"
	"github.com/antonpriyma/otus-highload/pkg/framework/grpc/interceptors/client"
	"github.com/antonpriyma/otus-highload/pkg/framework/service"
	"github.com/antonpriyma/otus-highload/pkg/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type AppConfig struct {
	service.Config `mapstructure:",squash"`
	UsersConfig    UsersConfig   `mapstructure:"users"`
	PostsConfig    PostsConfig   `mapstructure:"posts"`
	DialogsConfig  DialogsConfig `mapstructure:"dialogs"`
}

type DialogsConfig struct {
	Repo       dialog_repo.Config `mapstructure:"repository"`
	GRPCAddr   string             `mapstructure:"grpc_addr"`
	RabbitAddr string             `mapstructure:"rabbit_addr"`
}

type UsersConfig struct {
	Repo user_repo.Config `mapstructure:"repository"`
}

type PostsConfig struct {
	Repo post_repo.Config `mapstructure:"repository"`
}

func (a AppConfig) APIConfig() echoapi.Config {
	return echoapi.Config{
		ServeConfig: service.ServeConfig{
			GracefulWait: time.Second,
			StopWait:     time.Second,
		},
		Listen: ":8081",
	}
}

func main() {
	var cfg AppConfig
	cfg.Version = service.Version{
		Dist:    "local",
		Release: "local",
	}

	svc := echoapi.New(&cfg)

	userRepository, err := user_repo.NewUserRepository(cfg.UsersConfig.Repo, svc.Logger)
	utils.Must(svc.Logger, err, "failed to create users repository")

	sessionRepository := map_repository.NewSessionRepository(svc.Logger)

	usersUsecase := usecase.NewUserUsecase(userRepository, sessionRepository, svc.Logger)
	usersDelivery := user_delivery.NewUserDelivery(usersUsecase, svc.Logger)

	postRepository, err := post_repo.NewPostRepository(cfg.PostsConfig.Repo, svc.Logger)
	utils.Must(svc.Logger, err, "failed to create posts repository")

	conn, err := amqp.Dial(cfg.DialogsConfig.RabbitAddr)
	utils.Must(svc.Logger, err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	utils.Must(svc.Logger, err, "Failed to open a channel")
	defer ch.Close()

	notifier, err := notifer.NewNotifer(ch)
	utils.Must(svc.Logger, err, "Failed to create notifier")

	postUsecase := post_usecase.NewPostUsecase(postRepository, userRepository, notifier, svc.Logger)
	postDelivery := post_delivery.NewPostDelivery(postUsecase, svc.Logger)

	svc.API.Use(middleware.AuthMiddleware)
	svc.API.GET("/user/register", func(c echo.Context) error {
		req := models.User{
			ID:         models.UserID(uuid.New().String()),
			Username:   generateUsername(),
			FirstName:  "loadtest",
			SecondName: "loadtest",
			Biography:  generateBio(),
			Age:        123,
			Sex:        models.UserSex(generateSex()),
			City:       "Tver",
			Password:   generatePass(),
		}
		userID, err := usersDelivery.CreateUser(c.Request().Context(), req)
		if err != nil {
			return err
		}

		type RegisterResponse struct {
			UserID models.UserID `json:"user_id"`
		}

		return c.JSON(http.StatusOK, RegisterResponse{UserID: userID})
	})

	svc.API.POST("/login", func(c echo.Context) error {
		// TODO: login model
		req := new(user_delivery.UserLoginRequest)
		if err := c.Bind(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "failed to bind request").SetInternal(err)
		}
		token, err := usersDelivery.Login(c.Request().Context(), models.UserID(req.ID), req.Password)
		if err != nil {
			return err
		}

		type LoginResponse struct {
			Token string `json:"token"`
		}

		return c.JSON(http.StatusOK, LoginResponse{Token: string(token)})
	})

	svc.API.GET("/user/:id", func(c echo.Context) error {
		userID := c.Param("id")
		user, err := usersDelivery.GetUser(c.Request().Context(), models.UserID(userID))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, user)
	})

	svc.API.GET("/user/search", func(c echo.Context) error {
		firstName := c.QueryParam("first_name")
		secondName := c.QueryParam("second_name")

		users, err := usersDelivery.SearchUser(c.Request().Context(), firstName, secondName)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, users)
	})

	svc.API.GET("/friend/add/:id", func(c echo.Context) error {
		id := c.Param("id")

		err := usersDelivery.CreateFriend(c.Request().Context(), models.UserID(id))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, nil)
	})

	svc.API.GET("/post/feed", func(c echo.Context) error {
		userID, ok := contextlib.GetUserID(echoutils.MustGetContext(c))
		if !ok {
			return echoerrors.ValidationError(errors.New("user id not found"), "user id not found", echoerrors.ValidationErrorFields{})
		}

		limitRaw := c.QueryParam("limit")
		offsetRaw := c.QueryParam("offset")
		if limitRaw == "" {
			limitRaw = "-1"
		}

		if offsetRaw == "" {
			offsetRaw = "-1"
		}
		limit, err := strconv.Atoi(limitRaw)
		if err != nil {
			return echoerrors.ValidationError(err, "limit is not valid", echoerrors.ValidationErrorFields{})
		}

		offset, err := strconv.Atoi(offsetRaw)
		if err != nil {
			return echoerrors.ValidationError(err, "offset is not valid", echoerrors.ValidationErrorFields{})
		}

		posts, err := postDelivery.GetFeed(c.Request().Context(), userID, limit, offset)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, posts)
	})

	svc.API.POST("/post/create", func(c echo.Context) error {
		userID, ok := contextlib.GetUserID(echoutils.MustGetContext(c))
		if !ok {
			return echoerrors.ValidationError(errors.New("user id not found"), "user id not found", echoerrors.ValidationErrorFields{})
		}

		type CreatePostRequest struct {
			Text string `json:"text"`
		}

		req := new(CreatePostRequest)
		if err := c.Bind(req); err != nil {
			return echoerrors.ValidationError(err, "failed to bind request", echoerrors.ValidationErrorFields{})
		}

		postID, err := postDelivery.CreatePost(c.Request().Context(), models.Post{
			ID:     models.PostID(uuid.New().String()),
			Text:   req.Text,
			UserID: userID,
		})

		if err != nil {
			return err
		}

		type CreatePostResponse struct {
			PostID models.PostID `json:"post_id"`
		}

		return c.JSON(http.StatusOK, CreatePostResponse{
			PostID: postID,
		})
	})

	grpcConn, err := grpc.Dial(
		cfg.DialogsConfig.GRPCAddr,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(client.NewUnaryClientRequestIDInterceptor(func(ctx context.Context) string {
			reqID := reqid.GetRequestID(ctx)
			if reqID == "" {
				reqID = uuid.New().String()
			}

			return reqID
		})),
		grpc.WithUnaryInterceptor(client.NewUnaryClientLoggingInterceptor(svc.Logger, client.LogParams{
			Debug: true,
		})),
		grpc.WithUnaryInterceptor(client.NewUnaryClientStatInterceptor(client.StatConfig{Service: "dialogs"}, svc.StatRegistry)))
	utils.Must(svc.Logger, err, "failed to dial grpc dialogs")
	defer func() {
		if err := grpcConn.Close(); err != nil {
			log.Println(err)
		}
	}()

	dialogsClient := dialogs.NewDialogsClient(grpcConn)
	svc.API.POST("/dialog/:user_id/send", func(context echo.Context) error {
		friendID := context.Param("user_id")
		type SendRequest struct {
			Text string `json:"text"`
		}

		req := new(SendRequest)
		if err := context.Bind(req); err != nil {
			return echoerrors.ValidationError(err, "failed to bind request", echoerrors.ValidationErrorFields{})
		}

		userID, ok := contextlib.GetUserID(echoutils.MustGetContext(context))
		if !ok {
			return echoerrors.UnauthorizedError(errors.New("user id not found"), "user id not found", "user id not found")
		}

		_, err = dialogsClient.SendMessage(context.Request().Context(), &dialogs.SendMessageRequest{
			Message: &dialogs.Message{
				From: string(userID),
				To:   friendID,
				Text: req.Text,
			},
		})
		if err != nil {
			return err
		}

		return context.JSON(http.StatusOK, nil)
	})

	svc.API.GET("/dialog/:user_id/list", func(c echo.Context) error {
		friendID := c.Param("user_id")

		userID, ok := contextlib.GetUserID(echoutils.MustGetContext(c))
		if !ok {
			return echoerrors.UnauthorizedError(errors.New("user id not found"), "user id not found", "user id not found")
		}
		grpcMessages, err := dialogsClient.GetMessages(c.Request().Context(), &dialogs.GetMessagesRequest{
			From: friendID,
			User: string(userID),
		})
		if err != nil {
			return err
		}

		messages := make([]models.Message, 0, len(grpcMessages.Messages))
		for _, grpcMessage := range grpcMessages.Messages {
			messages = append(messages, models.Message{
				From: models.UserID(grpcMessage.From),
				To:   models.UserID(grpcMessage.To),
				Text: grpcMessage.Text,
			})
		}

		return c.JSON(http.StatusOK, messages)

	})

	svc.Run()
}

func generateSex() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(1)
}

func generatePass() string {
	rand.Seed(time.Now().UnixNano())
	characters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	username := make([]byte, 10)
	for i := range username {
		username[i] = characters[rand.Intn(len(characters))]
	}
	return string(username)
}

func generateUsername() string {
	rand.Seed(time.Now().UnixNano())
	characters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	username := make([]byte, 30)
	for i := range username {
		username[i] = characters[rand.Intn(len(characters))]
	}
	return string(username)
}

func generateBio() string {
	rand.Seed(time.Now().UnixNano())
	characters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	username := make([]byte, 50)
	for i := range username {
		username[i] = characters[rand.Intn(len(characters))]
	}
	return string(username)
}
