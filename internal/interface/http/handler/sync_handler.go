package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"im-chat/internal/domain/entity"
	"im-chat/internal/infra/database"
	"im-chat/internal/interface/http/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// syncCursor uses the database row ID as a tie breaker for legacy messages,
// which have sequence zero. The token is scoped by the query authorization.
type syncCursor struct {
	Seq uint64 `json:"s"`
	ID  uint   `json:"i"`
}

func encodeCursor(c syncCursor) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(raw string) (syncCursor, error) {
	var cursor syncCursor
	if raw == "" {
		return cursor, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(b) > 128 {
		return cursor, errors.New("invalid cursor")
	}
	if err := json.Unmarshal(b, &cursor); err != nil || cursor.ID == 0 {
		return cursor, errors.New("invalid cursor")
	}
	return cursor, nil
}

type syncOptions struct {
	limit     int
	cursor    syncCursor
	rawCursor string
	backwards bool
	tail      bool
}

func parseSyncOptions(c *gin.Context) (syncOptions, error) {
	options := syncOptions{limit: 50, tail: c.Query("tail") == "1"}
	if raw := c.Query("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			return options, errors.New("limit must be between 1 and 100")
		}
		options.limit = limit
	}
	after, before := c.Query("cursor"), c.Query("before")
	if (after != "" && before != "") || (options.tail && (after != "" || before != "")) {
		return options, errors.New("cursor, before and tail are mutually exclusive")
	}
	options.backwards = before != "" || options.tail
	options.rawCursor = after
	if before != "" {
		options.rawCursor = before
	}
	var err error
	options.cursor, err = decodeCursor(options.rawCursor)
	return options, err
}

func syncQuery(query *gorm.DB, options syncOptions) *gorm.DB {
	if options.rawCursor != "" {
		if options.backwards {
			query = query.Where("(seq < ? OR (seq = ? AND id < ?))", options.cursor.Seq, options.cursor.Seq, options.cursor.ID)
		} else {
			query = query.Where("(seq > ? OR (seq = ? AND id > ?))", options.cursor.Seq, options.cursor.Seq, options.cursor.ID)
		}
	}
	order := "seq ASC, id ASC"
	if options.backwards {
		order = "seq DESC, id DESC"
	}
	return query.Order(order).Limit(options.limit + 1)
}

type privateSyncMessage struct {
	*entity.PrivateMessage
	IsRead bool `json:"is_read"`
	Acked  bool `json:"acked"`
}

type groupSyncMessage struct {
	*entity.GroupMessage
	FromUsername string `json:"from_username"`
	FromNickname string `json:"from_nickname"`
	ReadCount    int64  `json:"read_count"`
	TotalCount   int64  `json:"total_count"`
	Acked        bool   `json:"acked"`
}

// SyncPrivate exposes only the authenticated user's own conversation.
func SyncPrivate(c *gin.Context) {
	uid, _ := c.Get("uid")
	reader, _ := uid.(string)
	peer := c.Query("friend_uid")
	options, err := parseSyncOptions(c)
	if err != nil || reader == "" || peer == "" || reader == peer {
		response.InvalidParams(c, "invalid conversation or cursor")
		return
	}
	var rows []entity.PrivateMessage
	query := database.DB.Where("(from_uid = ? AND to_uid = ?) OR (from_uid = ? AND to_uid = ?)", reader, peer, peer, reader)
	if err := syncQuery(query, options).Find(&rows).Error; err != nil {
		response.Failed(c, "获取消息失败")
		return
	}
	hasMore := len(rows) > options.limit
	if hasMore {
		rows = rows[:options.limit]
	}
	if options.backwards {
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
	}
	items := make([]privateSyncMessage, 0, len(rows))
	for i := range rows {
		items = append(items, privateSyncMessage{PrivateMessage: &rows[i], IsRead: rows[i].Read == 1, Acked: true})
	}
	first, last := syncCursor{}, syncCursor{}
	if len(rows) > 0 {
		first = syncCursor{Seq: rows[0].Seq, ID: rows[0].ID}
		last = syncCursor{Seq: rows[len(rows)-1].Seq, ID: rows[len(rows)-1].ID}
	}
	response.Success(c, syncPage(items, first, last, options, hasMore))
}

// SyncGroup verifies active membership on every request, including history.
func SyncGroup(c *gin.Context) {
	uid, _ := c.Get("uid")
	reader, _ := uid.(string)
	groupID := c.Query("group_id")
	options, err := parseSyncOptions(c)
	if err != nil || reader == "" || groupID == "" {
		response.InvalidParams(c, "invalid conversation or cursor")
		return
	}
	var member entity.GroupMember
	if err := database.DB.Where("group_id = ? AND uid = ? AND status = ?", groupID, reader, 1).First(&member).Error; err != nil {
		response.Forbidden(c)
		return
	}
	var rows []entity.GroupMessage
	if err := syncQuery(database.DB.Where("group_id = ?", groupID), options).Find(&rows).Error; err != nil {
		response.Failed(c, "获取群消息失败")
		return
	}
	hasMore := len(rows) > options.limit
	if hasMore {
		rows = rows[:options.limit]
	}
	if options.backwards {
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
	}
	items := make([]groupSyncMessage, 0, len(rows))
	if len(rows) > 0 {
		msgIDs := make([]string, 0, len(rows))
		uids := make([]string, 0, len(rows))
		for _, row := range rows {
			msgIDs = append(msgIDs, row.MsgID)
			uids = append(uids, row.FromUID)
		}
		var counts []struct {
			MsgID string
			Count int64
		}
		if err := database.DB.Model(&entity.MessageRead{}).Select("msg_id, COUNT(*) as count").Where("msg_id IN ?", msgIDs).Group("msg_id").Scan(&counts).Error; err != nil {
			response.Failed(c, "获取已读状态失败")
			return
		}
		var users []entity.User
		if err := database.DB.Where("uid IN ?", uids).Find(&users).Error; err != nil {
			response.Failed(c, "获取发送者失败")
			return
		}
		var memberCount int64
		if err := database.DB.Model(&entity.GroupMember{}).Where("group_id = ? AND status = ?", groupID, 1).Count(&memberCount).Error; err != nil {
			response.Failed(c, "获取群成员失败")
			return
		}
		countByID := make(map[string]int64, len(counts))
		for _, count := range counts {
			countByID[count.MsgID] = count.Count
		}
		usersByID := make(map[string]entity.User, len(users))
		for _, user := range users {
			usersByID[user.UID] = user
		}
		for i := range rows {
			user := usersByID[rows[i].FromUID]
			items = append(items, groupSyncMessage{GroupMessage: &rows[i], FromUsername: user.Username,
				FromNickname: user.Nickname, ReadCount: countByID[rows[i].MsgID], TotalCount: memberCount - 1, Acked: true})
		}
	}
	first, last := syncCursor{}, syncCursor{}
	if len(rows) > 0 {
		first = syncCursor{Seq: rows[0].Seq, ID: rows[0].ID}
		last = syncCursor{Seq: rows[len(rows)-1].Seq, ID: rows[len(rows)-1].ID}
	}
	response.Success(c, syncPage(items, first, last, options, hasMore))
}

func syncPage(messages interface{}, first, last syncCursor, options syncOptions, hasMore bool) gin.H {
	next := options.rawCursor
	older := ""
	if first.ID != 0 {
		older = encodeCursor(first)
		next = encodeCursor(last)
		if options.backwards && !options.tail {
			next = older
		}
	}
	if options.tail {
		return gin.H{"messages": messages, "next_cursor": next, "older_cursor": older, "has_more": false, "has_older": hasMore}
	}
	return gin.H{"messages": messages, "next_cursor": next, "older_cursor": older, "has_more": hasMore}
}
