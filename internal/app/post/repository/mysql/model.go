package mysql

import (
	"database/sql"

	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/go-sql-driver/mysql"
)

type Post struct {
	UUID   string `db:"uuid"`
	UserID string `db:"user_id"`
	Text   string `db:"text"`
}

func convertModelToPost(model models.Post) Post {
	return Post{
		UUID:   string(model.ID),
		UserID: string(model.UserID),
		Text:   model.Text,
	}
}

func convertPostToModel(post Post) models.Post {
	return models.Post{
		ID:     models.PostID(post.UUID),
		UserID: models.UserID(post.UserID),
		Text:   post.Text,
	}
}

func convertPostsToModels(posts []Post) []models.Post {
	res := make([]models.Post, 0, len(posts))
	for _, post := range posts {
		res = append(res, convertPostToModel(post))
	}

	return res
}

func convertSQLError(err error) error {
	if mysqlError, ok := err.(*mysql.MySQLError); ok {
		if mysqlError.Number == 1062 {
			return models.ErrUserAlreadyExists
		}
	}

	switch {
	case err == sql.ErrNoRows:
		return models.ErrPostNotFound
	}

	return err
}
