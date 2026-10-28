package websocket

import (
	"encoding/json"
	"errors"
	"testing"

	"im-chat/internal/infra/database"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupDelivery(t *testing.T) (sqlmock.Sqlmock, *Conn, *Conn) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	oldDB := database.DB
	database.DB = db
	sender := &Conn{UserID: "alice", SendChan: make(chan []byte, 2)}
	recipient := &Conn{UserID: "bob", SendChan: make(chan []byte, 2)}
	mu.Lock()
	userConns[sender.UserID] = sender
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		delete(userConns, sender.UserID)
		mu.Unlock()
		database.DB = oldDB
		sqlDB.Close()
	})
	return mock, sender, recipient
}

func expectMessageLookup(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT .* FROM `private_messages`").
		WithArgs("m1", "alice", "bob", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "msg_id", "from_uid", "to_uid"}).AddRow(7, "m1", "alice", "bob"))
}

func TestDeliveryAcknowledgementFollowsPersistence(t *testing.T) {
	mock, sender, recipient := setupDelivery(t)
	mock.ExpectBegin()
	expectMessageLookup(mock)
	mock.ExpectExec("UPDATE `private_messages`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	handleDeliveryReceipt(recipient, &WSMessage{MsgID: "m1", FromUID: "alice"})
	var notification map[string]interface{}
	select {
	case raw := <-sender.SendChan:
		if err := json.Unmarshal(raw, &notification); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatal("sender did not receive delivery notification")
	}
	if notification["type"] != "delivery_receipt" || notification["msg_id"] != "m1" || notification["delivered"] != true {
		t.Fatalf("incorrect delivery notification: %v", notification)
	}
	select {
	case raw := <-recipient.SendChan:
		var ack map[string]interface{}
		if err := json.Unmarshal(raw, &ack); err != nil {
			t.Fatal(err)
		}
		if ack["type"] != "delivery_ack" || ack["msg_id"] != "m1" {
			t.Fatalf("incorrect ack: %v", ack)
		}
	default:
		t.Fatal("recipient did not receive durable receipt acknowledgement")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeliveryFailureDoesNotAcknowledgeOrNotify(t *testing.T) {
	mock, sender, recipient := setupDelivery(t)
	mock.ExpectBegin()
	expectMessageLookup(mock)
	mock.ExpectExec("UPDATE `private_messages`").WillReturnError(errors.New("write failed"))
	mock.ExpectRollback()
	handleDeliveryReceipt(recipient, &WSMessage{MsgID: "m1", FromUID: "alice"})
	if len(sender.SendChan) != 0 || len(recipient.SendChan) != 0 {
		t.Fatal("a failed receipt was acknowledged or sent to the sender")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
