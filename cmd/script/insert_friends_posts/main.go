package main

import (
	"context"
	"encoding/csv"
	"github.com/antonpriyma/otus-highload/internal/app/user/repository/mysql"
	"math/rand"
	"os"
	"time"

	"github.com/antonpriyma/otus-highload/internal/app/models"
	mysql2 "github.com/antonpriyma/otus-highload/internal/app/post/repository/mysql"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/google/uuid"
)

const fileName = "cmd/script/insert_users/file.csv"

const n = 100

func main() {
	userRepository, err := mysql.NewUserRepository(mysql.Config{DataSourceName: "otus:otus@tcp(localhost:s)/otus"}, log.Default())
	if err != nil {
		panic(err)
	}

	postRepository, err := mysql2.NewPostRepository(mysql2.Config{DataSourceName: "otus:otus@tcp(localhost:s)/otus"}, log.Default())
	if err != nil {
		panic(err)
	}

	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	r := csv.NewReader(f)
	_, err = r.ReadAll()
	if err != nil {
		panic(err)
	}
	rand.Seed(time.Now().Unix())

	users, err := userRepository.GetRandomUsers(context.Background(), n)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(users); {
		user1 := users[rand.Intn(len(users))]

		var user2 models.User
		for {
			user2 = users[rand.Intn(len(users))]
			if err != nil {
				panic(err)
			}

			if user2.ID != user1.ID {
				break
			}
		}

		err = userRepository.CreateFriendship(context.Background(), user1.ID, user2.ID)
		if err != nil {
			continue
		}

		post1 := models.PostID(uuid.New().String())
		err = postRepository.CreatePost(context.Background(), models.Post{
			ID:     post1,
			UserID: user1.ID,
			Text:   generateRandomSentence(),
		})
		if err != nil {
			panic(err)
		}

		post2 := models.PostID(uuid.New().String())
		err = postRepository.CreatePost(context.Background(), models.Post{
			ID:     post2,
			UserID: user2.ID,
			Text:   generateRandomSentence(),
		})
		if err != nil {
			panic(err)
		}
		i++

		log.Default().Printf("user1: %s, user2: %s, post1: %s, post2: %s", user1.ID, user2.ID, post1, post2)
	}

	//var posts []models.Post
	//for i := 0; i < 1000; i++ {
	//	post := models.Post{
	//		ID:     models.PostID(uuid.New().String()),
	//		UserID: models.UserID("f7196faf-d8b5-461c-9a0b-b2f9a82f73cf"),
	//		Text:   generateRandomSentence(),
	//	}
	//	posts = append(posts, post)
	//}
	//
	//for _, post := range posts {
	//	err = postRepository.CreatePost(context.Background(), post)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
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
