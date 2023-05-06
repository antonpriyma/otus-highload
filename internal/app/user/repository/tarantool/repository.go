package tarantool

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/tarantool/go-tarantool"
	"strconv"
)

type userRepository struct {
	conn   *tarantool.Connection
	logger log.Logger
}

/*
users:format({
{name="uuid",type="uuid"},
{name="username",type="string"},
{name="first_name",type="string"},
{name="second_name",type="string"},
{name="age",type="integer"},
{name="sex",type="integer"},
{name="city",type="string"},
{name="biography",type="string"},
{name="password",type="string"},
})
*/

func (u userRepository) CreateUser(ctx context.Context, user models.User) error {
	req := map[string]interface{}{
		"uuid":        string(user.ID),
		"username":    user.Username,
		"first_name":  user.FirstName,
		"second_name": user.SecondName,
		"age":         user.Age,
		"sex":         user.Sex,
		"city":        user.City,
		"biography":   user.Biography,
		"password":    user.Password,
	}

	_, err := u.conn.Call("CreateUser", []interface{}{req})
	if err != nil {
		return err
	}
	return nil
}

func (u userRepository) GetAllUsersIDs(ctx context.Context) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (u userRepository) GetRandomUsers(ctx context.Context, n int) ([]models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u userRepository) GetFriends(ctx context.Context, userID models.UserID) ([]models.UserID, error) {
	//TODO implement me
	panic("implement me")
}

func (u userRepository) GetUser(ctx context.Context, userID models.UserID) (models.User, error) {
	resp, err := u.conn.Call("GetUser", []interface{}{
		map[string]interface{}{
			"uuid": string(userID),
		},
	})
	if err != nil {
		return models.User{}, err
	}

	var user1 models.User
	if len(resp.Data) > 0 {
		byteArr := []byte(fmt.Sprintf("%v", resp.Data[0]))
		type RawResp [][]interface{}
		rawResp := RawResp{}
		err = json.Unmarshal(byteArr, &rawResp)
		if err != nil {
			return models.User{}, err
		}

		age, _ := strconv.Atoi(fmt.Sprintf("%v", rawResp[0][4]))
		sex, _ := strconv.Atoi(fmt.Sprintf("%v", rawResp[0][5]))
		user1 = models.User{
			ID:         models.UserID(fmt.Sprintf("%v", rawResp[0][0])),
			Username:   fmt.Sprintf("%v", rawResp[0][1]),
			FirstName:  fmt.Sprintf("%v", rawResp[0][2]),
			SecondName: fmt.Sprintf("%v", rawResp[0][3]),
			Age:        age,
			Sex:        models.UserSex(sex),
			City:       fmt.Sprintf("%v", rawResp[0][6]),
			Biography:  fmt.Sprintf("%v", rawResp[0][7]),
			Password:   fmt.Sprintf("%v", rawResp[0][8]),
		}

	}

	return user1, nil
}

func (u userRepository) SearchUser(ctx context.Context, firstName string, secondName string) ([]models.User, error) {
	req := map[string]interface{}{
		"first_name":  firstName,
		"second_name": secondName,
	}

	resp, err := u.conn.Call("FindUser", []interface{}{req})
	if err != nil {
		return nil, err
	}

	var user1 models.User
	if len(resp.Data) > 0 {
		byteArr := []byte(fmt.Sprintf("%v", resp.Data[0]))
		type RawResp [][]interface{}
		rawResp := RawResp{}
		err = json.Unmarshal(byteArr, &rawResp)
		if err != nil {
			return nil, err
		}

		age, _ := strconv.Atoi(fmt.Sprintf("%v", rawResp[0][4]))
		sex, _ := strconv.Atoi(fmt.Sprintf("%v", rawResp[0][5]))
		user1 = models.User{
			ID:         models.UserID(fmt.Sprintf("%v", rawResp[0][0])),
			Username:   fmt.Sprintf("%v", rawResp[0][1]),
			FirstName:  fmt.Sprintf("%v", rawResp[0][2]),
			SecondName: fmt.Sprintf("%v", rawResp[0][3]),
			Age:        age,
			Sex:        models.UserSex(sex),
			City:       fmt.Sprintf("%v", rawResp[0][6]),
			Biography:  fmt.Sprintf("%v", rawResp[0][7]),
			Password:   fmt.Sprintf("%v", rawResp[0][8]),
		}

	}

	return []models.User{user1}, nil
}

func (u userRepository) CreateFriendship(ctx context.Context, userID1 models.UserID, userID2 models.UserID) error {
	//TODO implement me
	panic("implement me")
}

type Config struct {
	Host string
	User string
	Pass string
}

func NewUserRepository(cfg Config, logger log.Logger) (models.UserRepository, error) {
	conn, err := tarantool.Connect(cfg.Host, tarantool.Opts{
		User: cfg.User,
		Pass: cfg.Pass,
	})
	if err != nil {
		return nil, err
	}

	return userRepository{
		conn:   conn,
		logger: logger,
	}, nil
}
