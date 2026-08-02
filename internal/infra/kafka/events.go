package kafka

import "im-chat/internal/infra/config"

var topics *config.KafkaTopics

func InitTopics(cfg *config.KafkaConfig) {
	if cfg.Topics.GroupMsgRaw != "" {
		TopicGroupMsgRaw = cfg.Topics.GroupMsgRaw
	}
	if cfg.Topics.GroupMsgPublish != "" {
		TopicGroupMsgPublish = cfg.Topics.GroupMsgPublish
	}
	if cfg.Topics.PrivateMsg != "" {
		TopicPrivateMsg = cfg.Topics.PrivateMsg
	}
	if cfg.Topics.ReadEvents != "" {
		TopicReadEvents = cfg.Topics.ReadEvents
	}
}

var (
	TopicGroupMsgRaw     = "group-msg-raw"
	TopicGroupMsgPublish = "group-msg-publish"
	TopicPrivateMsg      = "private-msg"
	TopicReadEvents      = "read-events"
)

type GroupMessageEvent struct {
	MsgID       string `json:"msg_id"`
	ClientMsgID string `json:"client_msg_id,omitempty"`
	GroupID     string `json:"group_id"`
	FromUID     string `json:"from_uid"`
	Content     string `json:"content"`
	SendTime    int64  `json:"send_time"`
}

type PrivateMessageEvent struct {
	MsgID       string `json:"msg_id"`
	ClientMsgID string `json:"client_msg_id,omitempty"`
	FromUID     string `json:"from_uid"`
	ToUID       string `json:"to_uid"`
	Content     string `json:"content"`
	SendTime    int64  `json:"send_time"`
}

type ReadEvent struct {
	MsgID    string `json:"msg_id"`
	UID      string `json:"uid"`
	GroupID  string `json:"group_id,omitempty"`
	ReadTime int64  `json:"read_time"`
}

// BatchReadEvent 批量已读事件
type BatchReadEvent struct {
	UID      string                 `json:"uid"`
	GroupID  string                 `json:"group_id,omitempty"`
	ReadTime int64                  `json:"read_time"`
	Msgs     []ReadEventItem        `json:"msgs"`
}

// ReadEventItem 单条已读记录
type ReadEventItem struct {
	MsgID string `json:"msg_id"`
}
