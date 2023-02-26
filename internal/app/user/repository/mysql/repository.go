package mysql

import (
	"context"
	"database/sql"
	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"sync/atomic"
)

type userRepository struct {
	db           *sqlx.DB
	logger       log.Logger
	writeCounter atomic.Int32
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
	db, err := sqlx.Connect("mysql", "otus:Knowledge123_@tcp(localhost:3306)/otus")
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mysql")
	}

	return &userRepository{
		db:     db,
		logger: logger,
	}, nil
}

func (u *userRepository) CreateUser(ctx context.Context, model models.User) error {
	user := convertModelToUser(model)
	_, err := u.db.ExecContext(
		ctx,
		"INSERT INTO users (uuid, username, first_name, second_name, biography,age,sex,city,password) VALUES (UUID_TO_BIN(?),?,?,?,?,?,?,?,?)",
		user.UUID, user.Username, user.FirstName, user.SecondName, user.Biography, user.Age, user.Sex, user.City, user.Password,
	)

	if err != nil {
		return errors.Wrap(convertSQLError(err), "failed to insert into users")
	}

	u.writeCounter.Add(1)
	u.logger.Warn("write to db", "count", u.writeCounter.Load())
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
