package mysql

import (
	"context"
	"github.com/antonpriyma/otus-highload/internal/app/models"
	"github.com/antonpriyma/otus-highload/pkg/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DataSourceName string `mapstructure:"data_source_name"`
}
type repository struct {
	logger log.Logger
	db     *sqlx.DB
}

func convertSQLError(err error) error {
	return err
}

func (r repository) SendMessage(ctx context.Context, message models.Message) error {
	msg := convertModelToMessage(message)
	_, err := r.db.ExecContext(ctx, "INSERT INTO messages (sender_uuid, receiver_uuid, text) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?), (?))", msg.SenderUUID, msg.ReceiverUUID, msg.Text)
	if err != nil {
		return convertSQLError(err)
	}

	return nil
}

func (r repository) GetDialog(ctx context.Context, userID models.UserID, friendID models.UserID) ([]models.Message, error) {
	var messages []Message
	err := r.db.SelectContext(ctx, &messages, "SELECT BIN_TO_UUID(sender_uuid) as sender_uuid, BIN_TO_UUID(receiver_uuid) as receiver_uuid, text FROM messages WHERE (sender_uuid = UUID_TO_BIN(?) AND receiver_uuid = UUID_TO_BIN(?)) OR (sender_uuid = UUID_TO_BIN(?) AND receiver_uuid = UUID_TO_BIN(?))", userID, friendID, friendID, userID)
	if err != nil {
		return nil, convertSQLError(err)
	}

	return convertMessagesToModels(messages), nil
}

func NewRepository(cfg Config, logger log.Logger) (models.DialogRepository, error) {
	db, err := sqlx.Connect("mysql", cfg.DataSourceName)
	if err != nil {
		return nil, err
	}
	return repository{
		logger: logger,
		db:     db,
	}, nil
}
