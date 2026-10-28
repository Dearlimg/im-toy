package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"im-chat/internal/infra/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func deliveryTestDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB, SkipInitializeWithVersion: true,
	}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	previous := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = previous
		sqlDB.Close()
	})
	return mock
}

func deliveryContext(path, uid string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("GET", path, nil)
	c.Set("uid", uid)
	return c, recorder
}

func TestUndeliveredReturnsOnlyRecipientMessagesWithoutMarkingThemRead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := deliveryTestDB(t)
	mock.ExpectQuery("SELECT .* FROM `private_messages` WHERE .*to_uid = \\? AND `read` = \\? AND.*delivered.*ORDER BY id ASC").
		WithArgs("bob", 0, false, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "msg_id", "from_uid", "to_uid", "content", "read", "delivered"}).
			AddRow(1, "m1", "alice", "bob", "ciphertext", 0, false).
			AddRow(2, "m2", "alice", "bob", "later", 0, false))
	c, recorder := deliveryContext("/api/private/undelivered?limit=1", "bob")
	new(PrivateMessageHandler).GetUndelivered(c)
	var result struct {
		Code int `json:"code"`
		Data struct {
			Messages []struct {
				MsgID   string `json:"msg_id"`
				Content string `json:"content"`
			} `json:"messages"`
			HasMore bool `json:"has_more"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Code != 0 || len(result.Data.Messages) != 1 || result.Data.Messages[0].MsgID != "m1" || result.Data.Messages[0].Content != "ciphertext" || !result.Data.HasMore {
		t.Fatalf("unexpected result: %s", recorder.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPrivateStatusIsRestrictedToSenderAndReadImpliesDelivered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := deliveryTestDB(t)
	mock.ExpectQuery("SELECT .* FROM `private_messages` WHERE .*from_uid = \\? AND msg_id IN.*").
		WithArgs("alice", "m1", "not-alice").
		WillReturnRows(sqlmock.NewRows([]string{"msg_id", "delivered", "read"}).AddRow("m1", false, 1))
	c, recorder := deliveryContext("/api/private/status?msg_ids=m1,not-alice", "alice")
	new(PrivateMessageHandler).GetPrivateStatus(c)
	var result struct {
		Code int `json:"code"`
		Data struct {
			Messages []struct {
				MsgID     string `json:"msg_id"`
				Delivered bool   `json:"delivered"`
				IsRead    bool   `json:"is_read"`
			} `json:"messages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Code != 0 || len(result.Data.Messages) != 1 || result.Data.Messages[0].MsgID != "m1" || !result.Data.Messages[0].Delivered || !result.Data.Messages[0].IsRead {
		t.Fatalf("unexpected result: %s", recorder.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
