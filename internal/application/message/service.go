// Package message implements durable acceptance for private and group messages.
package message

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"im-chat/internal/domain/entity"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInvalid   = errors.New("invalid message")
	ErrForbidden = errors.New("message recipient or group is unavailable")
	ErrConflict  = errors.New("client message ID reused with different message")
)

type Input struct {
	Kind        string
	SenderUID   string
	TargetID    string
	ClientMsgID string
	Content     string
}

type Result struct {
	MsgID     string `json:"msg_id"`
	Seq       uint64 `json:"seq"`
	SendTime  int64  `json:"send_time"`
	Delivered bool   `json:"delivered"`
	IsRead    bool   `json:"is_read"`
	Duplicate bool   `json:"-"`
}

func Validate(in Input) error {
	if (in.Kind != "private" && in.Kind != "group") || in.SenderUID == "" || in.TargetID == "" {
		return ErrInvalid
	}
	if len(in.ClientMsgID) == 0 || len(in.ClientMsgID) > 64 || strings.TrimSpace(in.ClientMsgID) != in.ClientMsgID {
		return ErrInvalid
	}
	if in.Content == "" || utf8.RuneCountInString(in.Content) > 2000 || !utf8.ValidString(in.Content) {
		return ErrInvalid
	}
	if in.Kind == "private" && in.SenderUID == in.TargetID {
		return ErrInvalid
	}
	return nil
}

func digest(v any) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func conversationKey(in Input) string {
	if in.Kind == "group" {
		return digest([]string{in.Kind, in.TargetID})
	}
	a, b := in.SenderUID, in.TargetID
	if a > b {
		a, b = b, a
	}
	return digest([]string{in.Kind, a, b})
}

func fingerprint(in Input) string {
	return digest([]string{in.Kind, in.TargetID, in.Content})
}

func newMessageID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

func isDuplicate(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.Is(err, gorm.ErrDuplicatedKey) || (errors.As(err, &mysqlErr) && mysqlErr.Number == 1062)
}

// Send atomically records the request, allocates a conversation sequence and stores the message.
// A matching retry returns its original result without creating another message.
func Send(ctx context.Context, db *gorm.DB, in Input) (Result, error) {
	var result Result
	if err := Validate(in); err != nil {
		return result, err
	}
	if db == nil {
		return result, errors.New("database unavailable")
	}
	msgID, err := newMessageID()
	if err != nil {
		return result, fmt.Errorf("generate message ID: %w", err)
	}
	request := entity.MessageRequest{
		SenderUID: in.SenderUID, ClientMsgID: in.ClientMsgID,
		Fingerprint: fingerprint(in), MsgID: msgID, SendTime: time.Now().UnixMilli(),
	}
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Reserve the request key before allocating a sequence. Rollback releases
		// it if validation or storage fails, allowing the caller to retry safely.
		if err := tx.Create(&request).Error; err != nil {
			return err
		}
		if in.Kind == "group" {
			var member entity.GroupMember
			if err := tx.Where("group_id = ? AND uid = ? AND status = ?", in.TargetID, in.SenderUID, 1).First(&member).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrForbidden
				}
				return err
			}
		} else {
			var user entity.User
			if err := tx.Where("uid = ?", in.TargetID).First(&user).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrForbidden
				}
				return err
			}
		}
		key := conversationKey(in)
		counter := entity.MessageConversation{ConversationKey: key}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&counter).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("conversation_key = ?", key).First(&counter).Error; err != nil {
			return err
		}
		counter.LastSeq++
		if err := tx.Model(&counter).Update("last_seq", counter.LastSeq).Error; err != nil {
			return err
		}
		request.Seq = counter.LastSeq
		if err := tx.Model(&request).Update("seq", request.Seq).Error; err != nil {
			return err
		}
		if in.Kind == "private" {
			msg := entity.PrivateMessage{MsgID: msgID, Seq: request.Seq, FromUID: in.SenderUID, ToUID: in.TargetID, Content: in.Content, Type: 1, SendTime: request.SendTime}
			return tx.Create(&msg).Error
		}
		msg := entity.GroupMessage{MsgID: msgID, Seq: request.Seq, GroupID: in.TargetID, FromUID: in.SenderUID, Content: in.Content, Type: 1, SendTime: request.SendTime}
		return tx.Create(&msg).Error
	})
	if err != nil {
		if !isDuplicate(err) {
			return result, err
		}
		var stored entity.MessageRequest
		if lookupErr := db.WithContext(ctx).Where("sender_uid = ? AND client_msg_id = ?", in.SenderUID, in.ClientMsgID).First(&stored).Error; lookupErr != nil {
			return result, err
		}
		if stored.Fingerprint != request.Fingerprint {
			return result, ErrConflict
		}
		result := Result{MsgID: stored.MsgID, Seq: stored.Seq, SendTime: stored.SendTime, Duplicate: true}
		if in.Kind == "private" {
			var msg entity.PrivateMessage
			if lookupErr := db.WithContext(ctx).Where("msg_id = ? AND from_uid = ? AND to_uid = ?", stored.MsgID, in.SenderUID, in.TargetID).First(&msg).Error; lookupErr != nil {
				return Result{}, fmt.Errorf("load stored message: %w", lookupErr)
			}
			result.IsRead = msg.Read == 1
			result.Delivered = msg.Delivered || result.IsRead
		}
		return result, nil
	}
	return Result{MsgID: request.MsgID, Seq: request.Seq, SendTime: request.SendTime}, nil
}
