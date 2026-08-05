package message_repository

import (
	"context"
	"time"

	"github.com/evolution-foundation/evolution-go/pkg/config"
	message_model "github.com/evolution-foundation/evolution-go/pkg/message/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MessageRepository interface {
	InsertMessage(message message_model.Message) error
	GetMessageByID(messageID string) (*message_model.Message, error)
	GetMessageByIDContext(ctx context.Context, messageID string) (*message_model.Message, error)
	DeleteAllMessages() (int64, error)
	GetLatestMessageID(source string) (string, string, error)
}

type messageRepository struct {
	db           *gorm.DB
	queryTimeout time.Duration
}

func (m *messageRepository) dbWithTimeout(ctx context.Context) (*gorm.DB, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if m.queryTimeout <= 0 {
		return m.db.WithContext(ctx), func() {}
	}
	ctx, cancel := context.WithTimeout(ctx, m.queryTimeout)
	return m.db.WithContext(ctx), cancel
}

func messageUpdateColumns(message message_model.Message) []string {
	updates := []string{"timestamp", "status", "source"}
	if len(message.Referral) > 0 {
		updates = append(updates, "referral")
	}

	return updates
}

func (m *messageRepository) InsertMessage(message message_model.Message) error {
	db, cancel := m.dbWithTimeout(context.Background())
	defer cancel()

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "message_id"}},
		DoUpdates: clause.AssignmentColumns(messageUpdateColumns(message)),
	}).Create(&message).Error
}

func (m *messageRepository) GetMessageByID(messageID string) (*message_model.Message, error) {
	return m.GetMessageByIDContext(context.Background(), messageID)
}

func (m *messageRepository) GetMessageByIDContext(ctx context.Context, messageID string) (*message_model.Message, error) {
	db, cancel := m.dbWithTimeout(ctx)
	defer cancel()

	var message message_model.Message
	err := db.Where("message_id = ?", messageID).First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &message, nil
}

func (m *messageRepository) DeleteAllMessages() (int64, error) {
	db, cancel := m.dbWithTimeout(context.Background())
	defer cancel()

	result := db.Exec("DELETE FROM messages")
	return result.RowsAffected, result.Error
}

func (m *messageRepository) GetLatestMessageID(source string) (string, string, error) {
	db, cancel := m.dbWithTimeout(context.Background())
	defer cancel()

	var message message_model.Message
	err := db.Where("source = ?", source).Order("timestamp DESC").First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", "", nil
		}
		return "", "", err
	}

	return message.MessageID, message.Timestamp, nil
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return NewMessageRepositoryWithTimeout(db, config.DefaultDBQueryTimeout)
}

func NewMessageRepositoryWithTimeout(db *gorm.DB, queryTimeout time.Duration) MessageRepository {
	return &messageRepository{db: db, queryTimeout: queryTimeout}
}
