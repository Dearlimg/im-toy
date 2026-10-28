package message

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func mockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	return db, mock
}

func expectUser(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT .* FROM `users`").WillReturnRows(sqlmock.NewRows([]string{"id", "uid"}).AddRow(1, "bob"))
}

func TestSendCommitsBeforeSuccess(t *testing.T) {
	db, mock := mockDB(t)
	in := Input{Kind: "private", SenderUID: "alice", TargetID: "bob", ClientMsgID: "client-1", Content: "hello"}
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `message_requests`").WillReturnResult(sqlmock.NewResult(0, 1))
	expectUser(mock)
	mock.ExpectExec("INSERT INTO `message_conversations`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT .* FROM `message_conversations`").WillReturnRows(sqlmock.NewRows([]string{"conversation_key", "last_seq"}).AddRow(conversationKey(in), 4))
	mock.ExpectExec("UPDATE `message_conversations`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE `message_requests`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO `private_messages`").WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectCommit()
	got, err := Send(context.Background(), db, in)
	if err != nil || got.Seq != 5 || got.MsgID == "" || got.Duplicate {
		t.Fatalf("unexpected result %+v: %v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSendStorageFailureRollsBack(t *testing.T) {
	db, mock := mockDB(t)
	in := Input{Kind: "private", SenderUID: "alice", TargetID: "bob", ClientMsgID: "client-2", Content: "hello"}
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `message_requests`").WillReturnError(errors.New("database unavailable"))
	mock.ExpectRollback()
	got, err := Send(context.Background(), db, in)
	if err == nil || got.MsgID != "" {
		t.Fatalf("failed transaction returned success: %+v %v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSendRetryReturnsOriginalAndRejectsConflict(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		t.Run(map[bool]string{false: "same request", true: "changed content"}[conflict], func(t *testing.T) {
			db, mock := mockDB(t)
			in := Input{Kind: "private", SenderUID: "alice", TargetID: "bob", ClientMsgID: "client-3", Content: "hello"}
			fp := fingerprint(in)
			if conflict {
				fp = fingerprint(Input{Kind: "private", SenderUID: "alice", TargetID: "bob", Content: "different"})
			}
			mock.ExpectBegin()
			mock.ExpectExec("INSERT INTO `message_requests`").WillReturnError(&mysqlDriver.MySQLError{Number: 1062, Message: "duplicate"})
			mock.ExpectRollback()
			mock.ExpectQuery("SELECT .* FROM `message_requests`").WillReturnRows(sqlmock.NewRows([]string{"sender_uid", "client_msg_id", "fingerprint", "msg_id", "seq", "send_time"}).AddRow("alice", "client-3", fp, "original", 5, 1234))
			if !conflict {
				mock.ExpectQuery("SELECT .* FROM `private_messages`").
					WithArgs("original", "alice", "bob", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "msg_id", "delivered", "read"}).AddRow(1, "original", false, 1))
			}
			got, err := Send(context.Background(), db, in)
			if conflict {
				if !errors.Is(err, ErrConflict) {
					t.Fatalf("expected conflict, got %v", err)
				}
			} else if err != nil || !got.Duplicate || got.MsgID != "original" || got.Seq != 5 || !got.Delivered || !got.IsRead {
				t.Fatalf("unexpected retry result %+v: %v", got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
