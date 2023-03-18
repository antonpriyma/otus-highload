package http

import (
	"context"

	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/echo/echoerrors"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

type userDelivery struct {
	usecase models.UserUsecase
	logger  log.Logger
}

func (u userDelivery) CreateFriend(ctx context.Context, userID models.UserID) error {
	err := u.usecase.CreateFriend(ctx, userID)
	if err != nil {
		return errors.Wrap(convertUserError(err), "failed to create friendship")
	}
	return nil
}

func (u userDelivery) SearchUser(ctx context.Context, firstName string, secondName string) ([]models.User, error) {
	users, err := u.usecase.SearchUser(ctx, firstName, secondName)
	if err != nil {
		return nil, errors.Wrap(convertUserError(err), "failed to search user")
	}

	return users, nil
}

func NewUserDelivery(usecase models.UserUsecase, logger log.Logger) models.UserDelivery {
	return &userDelivery{
		usecase: usecase,
		logger:  logger,
	}
}

func (u userDelivery) CreateUser(ctx context.Context, user models.User) (models.UserID, error) {
	userID, err := u.usecase.CreateUser(ctx, user)
	if err != nil {
		return models.EmptyUserID, errors.Wrap(convertUserError(err), "failed to create user")
	}

	return userID, nil
}

func (u userDelivery) GetUser(ctx context.Context, userID models.UserID) (models.User, error) {
	user, err := u.usecase.GetUser(ctx, userID)
	if err != nil {
		return models.User{}, errors.Wrap(convertUserError(err), "failed to get user")
	}

	return user, nil
}

func (u userDelivery) Login(ctx context.Context, userID models.UserID, password string) (models.SessionToken, error) {
	token, err := u.usecase.CreateSession(ctx, userID, password)
	if err != nil {
		return models.EmptySessionToken, errors.Wrap(convertUserError(err), "failed to create session")
	}

	return token, nil
}

func convertUserError(err error) error {
	switch {
	case errors.Is(err, models.ErrUserAlreadyExists):
		return echoerrors.AlreadyExistsError(err, "username")
	case errors.Is(err, models.ErrWrongPassword):
		return echoerrors.UnauthorizedError(err, echoerrors.ReasonAuthError, "wrong password")
	case errors.Is(err, models.ErrUserNotFound):
		return echoerrors.NotFoundError(err, "user")
	default:
		return echoerrors.InternalError(err)
	}
}
