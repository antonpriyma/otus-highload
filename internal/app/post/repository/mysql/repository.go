package mysql

import (
	"context"
	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type postRepository struct {
	db     *sqlx.DB
	redis  *redis.Client
	logger log.Logger
}

func (p postRepository) CreatePost(ctx context.Context, model models.Post) error {
	post := convertModelToPost(model)
	_, err := p.db.ExecContext(ctx, "INSERT INTO post (uuid, user_id, text) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?), ?)", post.UUID, post.UserID, post.Text)
	if err != nil {
		return errors.Wrap(convertSQLError(err), "failed to create post")
	}

	return nil
}

type Config struct {
	DataSourceName string `mapstructure:"data_source_name"`
}

func NewPostRepository(cfg Config, logger log.Logger) (models.PostRepository, error) {
	db, err := sqlx.Connect("mysql", "otus:Knowledge123_@tcp(localhost:3306)/otus")
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mysql")
	}

	redis := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	res := redis.Ping(context.Background())
	if res.Err() != nil {
		return nil, errors.Wrap(res.Err(), "failed to connect to redis")
	}
	return postRepository{
		redis:  redis,
		db:     db,
		logger: logger,
	}, nil
}

func (p postRepository) GetFeed(ctx context.Context, userID string, limit int, offset int) ([]models.Post, error) {
	ok := p.redis.Exists(ctx, userID).Val()

	if ok == 1 && limit+offset < 1000 {
		var posts []models.Post
		var err error
		if limit != -1 && offset != -1 {
			err = p.redis.LRange(ctx, userID, int64(offset), int64(limit+offset)).ScanSlice(&posts)
		} else {
			err = p.redis.LRange(ctx, userID, 0, -1).ScanSlice(&posts)
		}

		if err != nil {
			return nil, err
		}

		return posts, nil
	}
	var posts []Post
	var err error
	if limit != -1 {
		err = p.db.SelectContext(ctx, &posts, "SELECT BIN_TO_UUID(uuid) as uuid, BIN_TO_UUID(user_id) as user_id, text FROM POST p INNER JOIN FRIENDS f ON ((p.user_id = f.user2  and f.user1 = UUID_TO_BIN(?))) or (p.user_id = f.user1  and f.user2 = UUID_TO_BIN(?)) LIMIT (?), (?)", userID, userID, offset, limit)
	} else {
		err = p.db.SelectContext(ctx, &posts, "SELECT BIN_TO_UUID(uuid) as uuid, BIN_TO_UUID(user_id) as user_id, text FROM POST p INNER JOIN FRIENDS f ON ((p.user_id = f.user2  and f.user1 = UUID_TO_BIN(?))) or (p.user_id = f.user1  and f.user2 = UUID_TO_BIN(?))", userID, userID)
	}
	if err != nil {
		return nil, err
	}

	return convertPostsToModels(posts), nil
}

func (p postRepository) AddToCache(ctx context.Context, userID string, post models.Post) error {
	marshalled, err := post.MarshalBinary()
	if err != nil {
		return err
	}
	err = p.redis.LPush(ctx, userID, marshalled).Err()
	if err != nil {
		return err
	}

	// TODO: make it configurable
	// Like lru cache
	err = p.redis.LTrim(ctx, userID, 0, 1000).Err()
	if err != nil {
		return err
	}

	return nil
}

func (p postRepository) GenerateCache(ctx context.Context, userID string) error {
	posts, err := p.GetFeed(ctx, userID, 0, 0)
	if err != nil {
		return err
	}

	for _, post := range posts {
		marshalled, err := post.MarshalBinary()
		if err != nil {
			return err
		}
		err = p.redis.LPush(ctx, userID, marshalled).Err()
		if err != nil {
			return err
		}
	}
	p.redis.LTrim(ctx, userID, 0, 1000).Err()
	if err != nil {
		return err
	}

	return nil
}
