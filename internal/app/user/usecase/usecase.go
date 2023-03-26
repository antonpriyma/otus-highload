package usecase

import (
	"context"

	"github.com/antonpriyma/otus-highload/internal/pkg/contextlib"
	"github.com/google/uuid"

	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	users    models.UserRepository
	sessions models.SessionRepository
	logger   log.Logger
}

func (u userUsecase) CreateFriend(ctx context.Context, userID models.UserID) error {
	ctxUserID, ok := contextlib.GetUserID(ctx)
	if !ok {
		return models.ErrUnauthorized
	}

	err := u.users.CreateFriendship(ctx, ctxUserID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to create friendship")
	}
	return nil
}

func (u userUsecase) SearchUser(ctx context.Context, firstName string, secondName string) ([]models.User, error) {
	users, err := u.users.SearchUser(ctx, firstName, secondName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search user")
	}

	return users, nil
}

func NewUserUsecase(users models.UserRepository, sessions models.SessionRepository, logger log.Logger) models.UserUsecase {
	return &userUsecase{
		users:    users,
		sessions: sessions,
		logger:   logger,
	}
}

func (u userUsecase) CreateUser(ctx context.Context, user models.User) (models.UserID, error) {
	pwd, err := hashAndSalt(user.Password)
	if err != nil {
		return models.EmptyUserID, errors.Wrap(err, "failed to hash password")
	}
	user.Password = pwd

	user.ID = models.UserID(uuid.New().String())

	err = u.users.CreateUser(ctx, user)
	if err != nil {
		return models.EmptyUserID, errors.Wrap(err, "failed to create user")
	}

	return user.ID, nil
}

func (u userUsecase) GetUser(ctx context.Context, userID models.UserID) (models.User, error) {
	user, err := u.users.GetUser(ctx, userID)
	if err != nil {
		return models.User{}, errors.Wrap(err, "failed to get user")
	}

	return user, nil
}

func (u userUsecase) CreateSession(ctx context.Context, userID models.UserID, password string) (models.SessionToken, error) {
	_, err := u.users.GetUser(ctx, userID)
	if err != nil {
		return models.EmptySessionToken, errors.Wrap(err, "failed to get user")
	}

	//if ok := comparePasswords(user.Password, password); !ok {
	//	return "", models.ErrWrongPassword
	//}

	sessionToken, err := u.sessions.CreateSession(ctx, userID)
	if err != nil {
		return "", errors.Wrap(err, "failed to create session")
	}

	return sessionToken, nil
}

func comparePasswords(hashedPassword string, plainPassword string) bool {
	byteHash := []byte(hashedPassword)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plainPassword))
	if err != nil {
		return false
	}

	return true
}

func hashAndSalt(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
