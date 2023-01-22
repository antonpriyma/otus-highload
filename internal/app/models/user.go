package models

import (
	"context"
	"encoding/json"

	"github.com/antonpriyma/otus-highload/pkg/errors"
)

var (
	ErrUserAlreadyExists = errors.Typed("user_already_exists", "user already exists")
	ErrUserNotFound      = errors.Typed("user_not_found", "user not found")
	ErrWrongPassword     = errors.Typed("wrong_password", "wrong password")
)

type UserID string

const (
	EmptyUserID UserID = ""
)

type UserSex int

const (
	UserSexMale UserSex = iota
	UserSexFemale
)

type SessionToken string

const (
	EmptySessionToken SessionToken = ""
)

type User struct {
	ID         UserID  `json:"id"`
	Username   string  `json:"username"`
	FirstName  string  `json:"first_name"`
	SecondName string  `json:"second_name"`
	Biography  string  `json:"biography"`
	Age        int     `json:"age"`
	Sex        UserSex `json:"sex"`
	City       string  `json:"city"`
	Password   string  `json:"password,omitempty"`
}

func (u User) MarshalJSON() ([]byte, error) {
	type user User // prevent recursion
	x := user(u)
	x.Password = ""
	return json.Marshal(x)
}

type UserDelivery interface {
	CreateUser(ctx context.Context, user User) (UserID, error)
	GetUser(ctx context.Context, userID UserID) (User, error)
	Login(ctx context.Context, userID UserID, password string) (SessionToken, error)
	SearchUser(ctx context.Context, firstName string, secondName string) ([]User, error)
}

type UserUsecase interface {
	CreateUser(ctx context.Context, user User) (UserID, error)
	GetUser(ctx context.Context, userID UserID) (User, error)
	CreateSession(ctx context.Context, userID UserID, password string) (SessionToken, error)
	SearchUser(ctx context.Context, firstName string, secondName string) ([]User, error)
}

type UserRepository interface {
	CreateUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, userID UserID) (User, error)
	SearchUser(ctx context.Context, firstName string, secondName string) ([]User, error)
}

type SessionRepository interface {
	CreateSession(ctx context.Context, userID UserID) (SessionToken, error)
}
