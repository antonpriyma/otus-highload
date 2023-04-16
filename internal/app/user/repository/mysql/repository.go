package mysql

import (
	"context"
	"database/sql"

	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db     *sqlx.DB
	logger log.Logger
}

func (u userRepository) GetAllUsersIDs(ctx context.Context) ([]string, error) {
	var userIDs []string
	err := u.db.SelectContext(ctx, &userIDs, "SELECT BIN_TO_UUID(uuid) as uuid FROM users")
	if err != nil {
		return nil, errors.Wrap(convertSQLError(err), "failed to get all users")
	}

	return userIDs, nil
}

func (u userRepository) GetFriends(ctx context.Context, userID models.UserID) ([]models.UserID, error) {
	var friends []Friendship
	err := u.db.SelectContext(ctx, &friends, "SELECT BIN_TO_UUID(user1) as user1, BIN_TO_UUID(user2) as user2 from friends where user1 = UUID_TO_BIN((?)) or UUID_TO_BIN((?))", userID, userID)
	if err != nil {
		return nil, errors.Wrap(convertSQLError(err), "failed to get friends")
	}

	var res []models.UserID
	for _, friend := range friends {
		if friend.User1 == string(userID) {
			res = append(res, models.UserID(friend.User2))
		} else {
			res = append(res, models.UserID(friend.User2))
		}
	}

	return res, nil
}

func (u userRepository) GetRandomUsers(ctx context.Context, n int) ([]models.User, error) {
	var user []User
	err := u.db.SelectContext(ctx, &user, "SELECT BIN_TO_UUID(uuid) as uuid, username, first_name, second_name, biography,age,sex,city,password FROM users ORDER BY RAND() LIMIT ?", n)
	if err != nil {
		return nil, errors.Wrap(convertSQLError(err), "failed to get user")
	}

	return convertUsersToModels(user), nil
}

func (u userRepository) CreateFriendship(ctx context.Context, userID1 models.UserID, userID2 models.UserID) error {
	_, err := u.db.ExecContext(ctx, "INSERT INTO friends (user1, user2) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?))", userID1, userID2)
	if err != nil {
		return errors.Wrap(convertSQLError(err), "failed to insert into friendships")
	}
	return nil
}

func (u userRepository) SearchUser(ctx context.Context, firstName string, secondName string) ([]models.User, error) {
	var res []User
	err := u.db.SelectContext(ctx, &res, "SELECT BIN_TO_UUID(uuid) as uuid, username, first_name, second_name, biography,age,sex,city,password FROM users WHERE first_name LIKE ? AND  second_name LIKE ?", firstName+"%", secondName+"%")
	if err != nil {
		return nil, errors.Wrap(convertSQLError(err), "failed to get users")
	}

	return convertUsersToModels(res), nil
}

type Config struct {
	DataSourceName string `mapstructure:"data_source_name"`
}

func NewUserRepository(cfg Config, logger log.Logger) (models.UserRepository, error) {
	db, err := sqlx.Connect("mysql", cfg.DataSourceName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mysql")
	}

	return userRepository{
		db:     db,
		logger: logger,
	}, nil
}

func (u userRepository) CreateUser(ctx context.Context, model models.User) error {
	user := convertModelToUser(model)
	_, err := u.db.ExecContext(
		ctx,
		"INSERT INTO users (uuid, username, first_name, second_name, biography,age,sex,city,password) VALUES (UUID_TO_BIN(?),?,?,?,?,?,?,?,?)",
		user.UUID, user.Username, user.FirstName, user.SecondName, user.Biography, user.Age, user.Sex, user.City, user.Password,
	)

	if err != nil {
		return errors.Wrap(convertSQLError(err), "failed to insert into users")
	}

	return nil
}

func (u userRepository) GetUser(ctx context.Context, userID models.UserID) (models.User, error) {
	res := User{}
	err := u.db.GetContext(ctx, &res, "SELECT BIN_TO_UUID(uuid) as uuid, username, first_name, second_name, biography,age,sex,city,password FROM users WHERE uuid = UUID_TO_BIN(?)", userID)
	if err != nil {
		return models.User{}, errors.Wrap(convertSQLError(err), "failed to get user")
	}

	return convertUserToModel(res), nil
}

func convertSQLError(err error) error {
	if mysqlError, ok := err.(*mysql.MySQLError); ok {
		if mysqlError.Number == 1062 {
			return models.ErrUserAlreadyExists
		}
	}

	switch {
	case err == sql.ErrNoRows:
		return models.ErrUserNotFound
	}

	return err
}
