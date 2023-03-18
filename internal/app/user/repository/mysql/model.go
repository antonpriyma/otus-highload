package mysql

import (
	"github.com/antonpriyma/otus-highload/internal/app/models"
)

type User struct {
	UUID       string `db:"uuid"`
	Username   string `db:"username"`
	FirstName  string `db:"first_name"`
	SecondName string `db:"second_name"`
	Biography  string `db:"biography"`
	Age        int    `db:"age"`
	Sex        int    `db:"sex"`
	City       string `db:"city"`
	Password   string `db:"password,omitempty"`
}

func convertModelToUser(model models.User) User {
	return User{
		UUID:       string(model.ID),
		Username:   model.Username,
		FirstName:  model.FirstName,
		SecondName: model.SecondName,
		Biography:  model.Biography,
		Age:        model.Age,
		Sex:        int(model.Sex),
		City:       model.City,
		Password:   model.Password,
	}
}

func convertUserToModel(user User) models.User {
	return models.User{
		ID:         models.UserID(user.UUID),
		Username:   user.Username,
		FirstName:  user.FirstName,
		SecondName: user.SecondName,
		Biography:  user.Biography,
		Age:        user.Age,
		City:       user.City,
		Password:   user.Password,
	}
}

func convertUsersToModels(users []User) []models.User {
	res := make([]models.User, 0, len(users))
	for _, user := range users {
		res = append(res, convertUserToModel(user))
	}

	return res
}

type Friendship struct {
	User1 string `db:"user1"`
	User2 string `db:"user2"`
}
