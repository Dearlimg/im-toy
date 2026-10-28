package readstate

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func testDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

func TestPrivateReadRequiresRecipient(t *testing.T) {
	db, mock := testDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `private_messages` WHERE .*msg_id = \\? AND from_uid = \\? AND to_uid = \\?.*").WithArgs("m1", "alice", "bob", 1).WillReturnRows(sqlmock.NewRows([]string{"id", "msg_id", "from_uid", "to_uid"}).AddRow(7, "m1", "alice", "bob"))
	mock.ExpectExec("UPDATE `private_messages` SET `delivered`=\\?,`read`=\\?,`updated_at`=\\?").
		WithArgs(true, 1, sqlmock.AnyArg(), 7, 0, false).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	msg, err := MarkPrivate(db, "bob", "alice", "m1")
	if err != nil || msg.FromUID != "alice" {
		t.Fatalf("valid read rejected: %+v, %v", msg, err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `private_messages` WHERE .*msg_id = \\? AND from_uid = \\? AND to_uid = \\?.*").WithArgs("m1", "alice", "mallory", 1).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()
	_, err = MarkPrivate(db, "mallory", "alice", "m1")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeliveryRequiresRecipientAndIsIdempotent(t *testing.T) {
	db, mock := testDB(t)
	for i := 0; i < 2; i++ {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT .* FROM `private_messages` WHERE .*msg_id = \\? AND from_uid = \\? AND to_uid = \\?.*").
			WithArgs("m1", "alice", "bob", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "msg_id", "from_uid", "to_uid"}).AddRow(7, "m1", "alice", "bob"))
		// A duplicate receipt legitimately updates zero rows but is still acknowledged.
		mock.ExpectExec("UPDATE `private_messages` SET .*`delivered`.*").
			WillReturnResult(sqlmock.NewResult(0, int64(1-i)))
		mock.ExpectCommit()
		msg, err := MarkDelivered(db, "bob", "alice", "m1")
		if err != nil || !msg.Delivered || msg.FromUID != "alice" {
			t.Fatalf("valid receipt rejected: %+v, %v", msg, err)
		}
	}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `private_messages` WHERE .*msg_id = \\? AND from_uid = \\? AND to_uid = \\?.*").
		WithArgs("m1", "alice", "mallory", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()
	if _, err := MarkDelivered(db, "mallory", "alice", "m1"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("non-recipient was allowed to mark a message delivered: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeliveryDatabaseFailureIsNotAcknowledged(t *testing.T) {
	db, mock := testDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `private_messages`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "msg_id", "from_uid", "to_uid"}).AddRow(7, "m1", "alice", "bob"))
	mock.ExpectExec("UPDATE `private_messages`").WillReturnError(errors.New("write failed"))
	mock.ExpectRollback()
	if msg, err := MarkDelivered(db, "bob", "alice", "m1"); err == nil || msg != nil {
		t.Fatalf("write failure was reported as success: %+v, %v", msg, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGroupReadRequiresMembershipAndMessageBelongsToGroup(t *testing.T) {
	db, mock := testDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `group_members`").WithArgs("group-1", "mallory", 1, 1).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()
	_, err := MarkGroup(db, "mallory", "group-1", []string{"m1"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("non-member accepted: %v", err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM `group_members`").WithArgs("group-1", "bob", 1, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "uid", "group_id"}).AddRow(1, "bob", "group-1"))
	mock.ExpectQuery("SELECT .* FROM `group_messages`").WithArgs("group-1", "other-group-message").WillReturnRows(sqlmock.NewRows([]string{"id", "msg_id"}))
	mock.ExpectRollback()
	_, err = MarkGroup(db, "bob", "group-1", []string{"other-group-message"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("foreign message accepted: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
