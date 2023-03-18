package http

import (
	"context"

	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/log"
)

type PostDelivery struct {
	Posts models.PostUsecase
	log   log.Logger
}

func NewPostDelivery(posts models.PostUsecase, log log.Logger) PostDelivery {
	return PostDelivery{
		Posts: posts,
		log:   log,
	}
}

func (p PostDelivery) GetFeed(ctx context.Context, userID models.UserID, limit int, offset int) ([]models.Post, error) {
	posts, err := p.Posts.GetFeed(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	return posts, nil
}
