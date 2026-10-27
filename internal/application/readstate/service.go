package readstate

import (
	"errors"
	"time"

	"im-chat/internal/domain/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrUnauthorized = errors.New("message or conversation is not accessible")

// MarkPrivate records receipt only for a message addressed to reader. It never
// advances the sender's read state as a side effect of fetching history.
func MarkPrivate(db *gorm.DB, reader, peer, msgID string) (*entity.PrivateMessage, error) {
	if reader == "" || peer == "" || msgID == "" || reader == peer {
		return nil, ErrUnauthorized
	}
	var msg entity.PrivateMessage
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("msg_id = ? AND from_uid = ? AND to_uid = ?", msgID, peer, reader).First(&msg).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUnauthorized
			}
			return err
		}
		return tx.Model(&entity.PrivateMessage{}).Where("id = ? AND `read` = ?", msg.ID, 0).Update("read", 1).Error
	})
	return &msg, err
}

// MarkGroup persists a batch only after every ID is validated against the
// requested group and the reader is an active member.
func MarkGroup(db *gorm.DB, reader, groupID string, msgIDs []string) ([]entity.GroupMessage, error) {
	if reader == "" || groupID == "" || len(msgIDs) == 0 || len(msgIDs) > 200 {
		return nil, ErrUnauthorized
	}
	unique := make(map[string]struct{}, len(msgIDs))
	for _, id := range msgIDs {
		if id == "" {
			return nil, ErrUnauthorized
		}
		unique[id] = struct{}{}
	}
	var messages []entity.GroupMessage
	err := db.Transaction(func(tx *gorm.DB) error {
		var member entity.GroupMember
		if err := tx.Where("group_id = ? AND uid = ? AND status = ?", groupID, reader, 1).First(&member).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUnauthorized
			}
			return err
		}
		if err := tx.Where("group_id = ? AND msg_id IN ?", groupID, msgIDs).Find(&messages).Error; err != nil {
			return err
		}
		if len(messages) != len(unique) {
			return ErrUnauthorized
		}
		reads := make([]entity.MessageRead, 0, len(messages))
		now := time.Now().UnixMilli()
		for _, msg := range messages {
			if msg.FromUID != reader {
				reads = append(reads, entity.MessageRead{MsgID: msg.MsgID, UID: reader, ReadTime: now})
			}
		}
		if len(reads) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&reads).Error
	})
	return messages, err
}
