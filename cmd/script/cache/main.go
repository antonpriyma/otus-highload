package main

import (
	"context"
	"github.com/antonpriyma/otus-highload/internal/app/user/repository/mysql"
	"math/rand"
	"time"

	mysql2 "github.com/antonpriyma/otus-highload/internal/app/post/repository/mysql"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

func main() {
	postRepository, err := mysql2.NewPostRepository(mysql2.Config{DataSourceName: "otus:otus@tcp(localhost:s)/otus"}, log.Default())

	userRepository, err := mysql.NewUserRepository(mysql.Config{DataSourceName: "otus:otus@tcp(localhost:s)/otus"}, log.Default())
	if err != nil {
		panic(err)
	}

	IDs, err := userRepository.GetAllUsersIDs(context.Background())
	if err != nil {
		panic(err)
	}

	log.Default().Info("Start generating cache")

	for i, id := range IDs {
		log.Default().Infof("Generate cache for user %d/%d", i, len(IDs))
		err = postRepository.GenerateCache(context.Background(), id)
		if err != nil {
			panic(err)
		}
	}
}

func generateRandomSentence() string {
	rand.Seed(time.Now().UnixNano())
	characters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	username := make([]byte, 100)
	for i := range username {
		username[i] = characters[rand.Intn(len(characters))]
	}
	return string(username)
}
