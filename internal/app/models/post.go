package models

import (
	"context"
	"encoding/json"
)

type PostDelivery interface {
	GetFeed(ctx context.Context, userID UserID) ([]Post, error)
}

type PostUsecase interface {
	GetFeed(ctx context.Context, userID UserID, limit int, offset int) ([]Post, error)
}

type PostRepository interface {
	GetFeed(ctx context.Context, userID string, limit int, offset int) ([]Post, error)
	CreatePost(ctx context.Context, post Post) error
	GenerateCache(ctx context.Context, userID string) error
	AddToCache(ctx context.Context, userID string, post Post) error
}

type PostID string

type Post struct {
	ID     PostID `json:"id"`
	UserID UserID `json:"user_id"`
	Text   string `json:"text"`
}

func (p Post) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

func (p *Post) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}
