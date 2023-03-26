package main

import (
	"fmt"
	post_delivery "github.com/antonpriyma/otus-highload/internal/app/post/delivery/http"
	post_repo "github.com/antonpriyma/otus-highload/internal/app/post/repository/mysql"
	post_usecase "github.com/antonpriyma/otus-highload/internal/app/post/usecase"
	"github.com/antonpriyma/otus-highload/internal/pkg/contextlib"
	"github.com/antonpriyma/otus-highload/internal/pkg/middleware"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoerrors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoutils"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/antonpriyma/otus-highload/internal/app/models"
	map_repository "github.com/antonpriyma/otus-highload/internal/app/session/repository/map"
	user_delivery "github.com/antonpriyma/otus-highload/internal/app/user/delivery/http"
	user_repo "github.com/antonpriyma/otus-highload/internal/app/user/repository/mysql"
	"github.com/antonpriyma/otus-highload/internal/app/user/usecase"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoapi"
	"github.com/antonpriyma/otus-highload/pkg/framework/service"
	"github.com/antonpriyma/otus-highload/pkg/utils"

	"github.com/labstack/echo"
)

type AppConfig struct {
	service.Config `mapstructure:",squash"`
	UsersConfig    UsersConfig `mapstructure:"users"`
	PostsConfig    PostsConfig `mapstructure:"posts"`
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

	var count atomic.Int32

	svc := echoapi.New(&cfg)

	userRepository, err := user_repo.NewUserRepository(cfg.UsersConfig.Repo, svc.Logger)
	utils.Must(svc.Logger, err, "failed to create users repository")

	sessionRepository := map_repository.NewSessionRepository(svc.Logger)

	usersUsecase := usecase.NewUserUsecase(userRepository, sessionRepository, svc.Logger)
	usersDelivery := user_delivery.NewUserDelivery(usersUsecase, svc.Logger)

	postRepository, err := post_repo.NewPostRepository(cfg.PostsConfig.Repo, svc.Logger)
	utils.Must(svc.Logger, err, "failed to create posts repository")

	postUsecase := post_usecase.NewPostUsecase(postRepository, svc.Logger)
	postDelivery := post_delivery.NewPostDelivery(postUsecase, svc.Logger)

	svc.API.Use(middleware.AuthMiddleware)
	svc.API.POST("/user/register", func(c echo.Context) error {
		req := new(models.User)
		if err := c.Bind(req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "failed to bind request").SetInternal(err)
		}
		userID, err := usersDelivery.CreateUser(c.Request().Context(), *req)
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
		count.Add(1)
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

	svc.Run()
	defer fmt.Print(count)
}
