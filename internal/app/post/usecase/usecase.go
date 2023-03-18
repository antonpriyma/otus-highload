package usecase

import (
	"context"
	"github.com/antonpriyma/otus-highload/pkg/errors"

	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

type postUsecase struct {
	posts  models.PostRepository
	users  models.UserRepository
	logger log.Logger
}

func (p postUsecase) CreatePost(ctx context.Context, post models.Post) error {
	err := p.posts.CreatePost(ctx, post)
	if err != nil {
		return errors.Wrap(err, "failed to create post")
	}

	friendsList, err := p.users.GetFriends(ctx, post.UserID)
	if err != nil {
		return errors.Wrap(err, "failed to get friends list")
	}

	for _, friend := range friendsList {
		err = p.posts.AddToCache(ctx, string(friend), post)
		if err != nil {
			return errors.Wrap(err, "failed to add to cache")
		}
	}

	return nil
}

func (p postUsecase) GetFeed(ctx context.Context, userID models.UserID, limit int, offset int) ([]models.Post, error) {
	posts, err := p.posts.GetFeed(ctx, string(userID), limit, offset)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func NewPostUsecase(posts models.PostRepository, logger log.Logger) models.PostUsecase {
	return postUsecase{
		posts:  posts,
		logger: logger,
	}
}
