package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/internal/app/user/repository/mysql"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/google/uuid"
)

const fileName = "cmd/script/insert_users/file.csv"

func main() {
	repository, err := mysql.NewUserRepository(mysql.Config{DataSourceName: "otus:otus@tcp(localhost:3306)/otus"}, log.Default())
	if err != nil {
		panic(err)
	}

	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}

	for i, record := range records {
		if i%10000 == 0 {
			fmt.Printf("run %d/%d", i, len(records))
		}
		name := record[0]
		age := record[1]
		ageInt, _ := strconv.Atoi(age)
		city := record[2]

		splitted := strings.Split(name, " ")
		firstName, secondName := splitted[0], splitted[1]

		err := repository.CreateUser(context.Background(), models.User{
			ID:         models.UserID(uuid.New().String()),
			Username:   generateUsername(),
			FirstName:  firstName,
			SecondName: secondName,
			Biography:  generateBio(),
			Age:        ageInt,
			Sex:        models.UserSex(generateSex()),
			City:       city,
			Password:   generatePass(),
		})

		//res, err := http.Post("http://:8081/user/register", "application/json", strings.NewReader(string(body)))
		if err != nil {
			fmt.Printf("error: %s", err)
			continue
		}

		//if res.StatusCode != http.StatusOK {
		//	panic(errors.New("response code is not 200"))
		//}
	}
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
