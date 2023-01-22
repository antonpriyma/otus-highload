package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/antonpriyma/otus-highload/internal/app/models"
	map_repository "github.com/antonpriyma/otus-highload/internal/app/session/repository/map"
	user_delivery "github.com/antonpriyma/otus-highload/internal/app/user/delivery/http"
	"github.com/antonpriyma/otus-highload/internal/app/user/repository/mysql"
	"github.com/antonpriyma/otus-highload/internal/app/user/usecase"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoapi"
	"github.com/antonpriyma/otus-highload/pkg/framework/service"
	"github.com/antonpriyma/otus-highload/pkg/utils"

	"github.com/labstack/echo"
)

type AppConfig struct {
	service.Config `mapstructure:",squash"`
	UsersConfig    UsersConfig `mapstructure:"users"`
}

type UsersConfig struct {
	Repo mysql.Config `mapstructure:"repository"`
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

	userRepository, err := mysql.NewUserRepository(cfg.UsersConfig.Repo, svc.Logger)
	utils.Must(svc.Logger, err, "failed to create users repository")

	sessionRepository := map_repository.NewSessionRepository(svc.Logger)

	usersUsecase := usecase.NewUserUsecase(userRepository, sessionRepository, svc.Logger)
	usersDelivery := user_delivery.NewUserDelivery(usersUsecase, svc.Logger)

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

	svc.Run()
	defer fmt.Print(count)
}
