package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type redisCache struct {
	client *redis.Client
}

func newRedisCache(client *redis.Client) *redisCache {
	return &redisCache{client: client}
}

func GetCtx() context.Context {
	return ctx
}

type ConnInfo struct {
	UID     string `json:"uid"`
	Server  string `json:"server"`
	RoomID  string `json:"room_id"`
	LoginAt int64  `json:"login_at"`
}

type UserBasicInfo struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type GroupBasicInfo struct {
	GroupID     string `json:"group_id"`
	Name        string `json:"name"`
	OwnerUID    string `json:"owner_uid"`
	Avatar      string `json:"avatar"`
	MaxMember   int    `json:"max_member"`
	MemberCount int    `json:"member_count"`
}

func (c *redisCache) SetUserOnline(uid string) error {
	key := fmt.Sprintf("im:online:%s", uid)
	return c.client.Set(ctx, key, "1", 60*time.Second).Err()
}

func (c *redisCache) SetUserOffline(uid string) error {
	key := fmt.Sprintf("im:online:%s", uid)
	return c.client.Del(ctx, key).Err()
}

func (c *redisCache) IsUserOnline(uid string) (bool, error) {
	key := fmt.Sprintf("im:online:%s", uid)
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (c *redisCache) RefreshUserOnline(uid string) error {
	key := fmt.Sprintf("im:online:%s", uid)
	return c.client.Expire(ctx, key, 60*time.Second).Err()
}

func (c *redisCache) SetConn(uid string, info *ConnInfo) error {
	key := fmt.Sprintf("im:conn:%s", uid)
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, 0).Err()
}

func (c *redisCache) GetConn(uid string) (*ConnInfo, error) {
	key := fmt.Sprintf("im:conn:%s", uid)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var info ConnInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *redisCache) RemoveConn(uid string) error {
	key := fmt.Sprintf("im:conn:%s", uid)
	return c.client.Del(ctx, key).Err()
}

func (c *redisCache) AddGroupMember(groupID, uid string) error {
	key := fmt.Sprintf("im:group:members:%s", groupID)
	return c.client.SAdd(ctx, key, uid).Err()
}

func (c *redisCache) RemoveGroupMember(groupID, uid string) error {
	key := fmt.Sprintf("im:group:members:%s", groupID)
	return c.client.SRem(ctx, key, uid).Err()
}

func (c *redisCache) GetGroupMembers(groupID string) ([]string, error) {
	key := fmt.Sprintf("im:group:members:%s", groupID)
	return c.client.SMembers(ctx, key).Result()
}

func (c *redisCache) GetGroupMemberCount(groupID string) (int64, error) {
	key := fmt.Sprintf("im:group:members:%s", groupID)
	return c.client.SCard(ctx, key).Result()
}

func (c *redisCache) IsGroupMember(groupID, uid string) (bool, error) {
	key := fmt.Sprintf("im:group:members:%s", groupID)
	return c.client.SIsMember(ctx, key, uid).Result()
}

func (c *redisCache) SetGroupMembers(groupID string, members []string) error {
	key := fmt.Sprintf("im:group:members:%s", groupID)
	pipe := c.client.Pipeline()
	pipe.Del(ctx, key)
	if len(members) > 0 {
		pipe.SAdd(ctx, key, convertToInterfaceSlice(members)...)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func convertToInterfaceSlice(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}

func (c *redisCache) SetUserInfo(uid string, info *UserBasicInfo) error {
	key := fmt.Sprintf("im:user:info:%s", uid)
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (c *redisCache) GetUserInfo(uid string) (*UserBasicInfo, error) {
	key := fmt.Sprintf("im:user:info:%s", uid)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var info UserBasicInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *redisCache) SetGroupInfo(groupID string, info *GroupBasicInfo) error {
	key := fmt.Sprintf("im:group:info:%s", groupID)
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (c *redisCache) GetGroupInfo(groupID string) (*GroupBasicInfo, error) {
	key := fmt.Sprintf("im:group:info:%s", groupID)
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var info GroupBasicInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *redisCache) IncrMsgLimit(uid string) (int64, error) {
	key := fmt.Sprintf("im:limiter:msg:%s", uid)
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		c.client.Expire(ctx, key, 1*time.Second)
	}
	return count, nil
}

func (c *redisCache) CheckAndSetIdempotentKey(uid, clientMsgID string) (bool, error) {
	key := fmt.Sprintf("im:idempotent:%s:%s", uid, clientMsgID)
	added, err := c.client.SetNX(ctx, key, "1", 24*time.Hour).Result()
	if err != nil {
		return false, err
	}
	return added, nil
}

func (c *redisCache) GetMsgLimit(uid string) (int64, error) {
	key := fmt.Sprintf("im:limiter:msg:%s", uid)
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

func (c *redisCache) AddMsgRead(msgID, uid string) error {
	key := fmt.Sprintf("im:msg:read:%s", msgID)
	return c.client.SAdd(ctx, key, uid).Err()
}

func (c *redisCache) GetMsgReadCount(msgID string) (int64, error) {
	key := fmt.Sprintf("im:msg:read:%s", msgID)
	return c.client.SCard(ctx, key).Result()
}

func (c *redisCache) BatchGetMsgReadCount(msgIDs []string) (map[string]int64, error) {
	if len(msgIDs) == 0 {
		return make(map[string]int64), nil
	}
	pipe := c.client.Pipeline()
	cmds := make(map[string]*redis.IntCmd)
	for _, msgID := range msgIDs {
		key := fmt.Sprintf("im:msg:read:%s", msgID)
		cmds[msgID] = pipe.SCard(ctx, key)
	}
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}
	result := make(map[string]int64)
	for msgID, cmd := range cmds {
		result[msgID], _ = cmd.Result()
	}
	return result, nil
}

func (c *redisCache) GetMsgReaders(msgID string) ([]string, error) {
	key := fmt.Sprintf("im:msg:read:%s", msgID)
	return c.client.SMembers(ctx, key).Result()
}

func (c *redisCache) AddMsgReadPending(msgID string, timestamp int64) error {
	key := "im:msg:read:pending"
	return c.client.ZAdd(ctx, key, redis.Z{Score: float64(timestamp), Member: msgID}).Err()
}

func (c *redisCache) PopMsgReadPending(count int64) ([]string, error) {
	key := "im:msg:read:pending"
	result, err := c.client.ZPopMin(ctx, key, count).Result()
	if err != nil {
		return nil, err
	}
	msgIDs := make([]string, 0, len(result))
	for _, z := range result {
		if msgID, ok := z.Member.(string); ok {
			msgIDs = append(msgIDs, msgID)
		}
	}
	return msgIDs, nil
}

func (c *redisCache) RemoveMsgReadPending(msgIDs []string) error {
	if len(msgIDs) == 0 {
		return nil
	}
	key := "im:msg:read:pending"
	Members := make([]interface{}, len(msgIDs))
	for i, id := range msgIDs {
		Members[i] = id
	}
	return c.client.ZRem(ctx, key, Members...).Err()
}

func (c *redisCache) Set(key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *redisCache) Get(key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *redisCache) Del(key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *redisCache) DelByPattern(pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			c.client.Del(ctx, keys...)
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
