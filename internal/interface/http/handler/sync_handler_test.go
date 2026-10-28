package handler

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"

	"im-chat/internal/infra/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestPrivateSyncTailAndCatchup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	previous := database.DB
	database.DB = db
	defer func() { database.DB = previous }()

	columns := []string{"id", "msg_id", "seq", "from_uid", "to_uid", "content", "send_time", "read", "delivered"}
	mock.ExpectQuery("SELECT .* FROM `private_messages`.*ORDER BY seq DESC, id DESC").
		WithArgs("alice", "bob", "bob", "alice", 3).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(3, "m2", 2, "alice", "bob", "two", 2, 1, false).
			AddRow(2, "m1", 1, "bob", "alice", "one", 1, 0, false).
			AddRow(1, "legacy", 0, "bob", "alice", "legacy", 0, 0, false))
	page := runPrivateSync(t, "/?friend_uid=bob&limit=2&tail=1")
	if len(page.Messages) != 2 || page.Messages[0].MsgID != "m1" || page.Messages[1].MsgID != "m2" || !page.HasOlder || page.NextCursor == "" {
		t.Fatalf("unexpected tail: %+v", page)
	}
	if page.Messages[0].Delivered || !page.Messages[1].Delivered || !page.Messages[1].IsRead {
		t.Fatalf("legacy read must imply delivery without claiming unread messages are delivered: %+v", page.Messages)
	}
	olderCursor := page.OlderCursor
	// A forward catch-up begins after the exact server cursor, not a guessed ID.
	mock.ExpectQuery("SELECT .* FROM `private_messages`.*seq > .*ORDER BY seq ASC, id ASC").
		WithArgs("alice", "bob", "bob", "alice", 2, 2, 3, 3).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(4, "m3", 3, "alice", "bob", "three", 3, 0, true))
	page = runPrivateSync(t, "/?friend_uid=bob&limit=2&cursor="+url.QueryEscape(page.NextCursor))
	if len(page.Messages) != 1 || page.Messages[0].MsgID != "m3" || page.HasMore {
		t.Fatalf("unexpected catch-up: %+v", page)
	}
	mock.ExpectQuery("SELECT .* FROM `private_messages`.*seq < .*ORDER BY seq DESC, id DESC").
		WithArgs("alice", "bob", "bob", "alice", 1, 1, 2, 3).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "legacy", 0, "bob", "alice", "legacy", 0, 0, false))
	page = runPrivateSync(t, "/?friend_uid=bob&limit=2&before="+url.QueryEscape(olderCursor))
	if len(page.Messages) != 1 || page.Messages[0].MsgID != "legacy" || page.HasMore {
		t.Fatalf("unexpected older page: %+v", page)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type testSyncPage struct {
	Messages []struct {
		MsgID     string `json:"msg_id"`
		Delivered bool   `json:"delivered"`
		IsRead    bool   `json:"is_read"`
	} `json:"messages"`
	NextCursor  string `json:"next_cursor"`
	OlderCursor string `json:"older_cursor"`
	HasOlder    bool   `json:"has_older"`
	HasMore     bool   `json:"has_more"`
}

func runPrivateSync(t *testing.T, path string) testSyncPage {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", path, nil)
	c.Set("uid", "alice")
	SyncPrivate(c)
	var result struct {
		Code int          `json:"code"`
		Data testSyncPage `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Code != 0 {
		t.Fatalf("sync returned %s", w.Body.String())
	}
	return result.Data
}
