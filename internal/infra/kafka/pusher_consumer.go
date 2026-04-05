package kafka

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"im-chat/internal/infra/redis"
	infra_repo "im-chat/internal/infra/repository"

	"github.com/IBM/sarama"
)

var globalPusher Pusher

func RegisterPusher(p Pusher) {
	globalPusher = p
}

type PusherConsumer struct {
	reader  sarama.ConsumerGroup
	topics  []string
	handler *pusherHandler
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

type pusherHandler struct {
	ready chan bool
}

func NewPusherConsumer(brokers []string, groupID string, topics []string) (*PusherConsumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	reader, err := sarama.NewConsumerGroup(brokers, groupID, saramaConfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &PusherConsumer{
		reader:  reader,
		topics:  topics,
		handler: &pusherHandler{ready: make(chan bool)},
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (c *PusherConsumer) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.reader.Consume(c.ctx, c.topics, c.handler); err != nil {
				log.Printf("PusherConsumer error: %v", err)
			}
			if c.ctx.Err() != nil {
				return
			}
			c.handler.ready = make(chan bool)
		}
	}()

	<-c.handler.ready
	log.Println("PusherConsumer started")
}

func (c *PusherConsumer) Stop() error {
	c.cancel()
	c.wg.Wait()
	return c.reader.Close()
}

func (h *pusherHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *pusherHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *pusherHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			h.processMessage(session, msg)
		case <-session.Context().Done():
			return nil
		}
	}
}

func (h *pusherHandler) processMessage(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	switch msg.Topic {
	case TopicGroupMsgPublish:
		var event GroupMessageEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal group message: %v", err)
			return
		}
		h.pushGroupMessage(&event)

	case TopicPrivateMsg:
		var event PrivateMessageEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal private message: %v", err)
			return
		}
		h.pushPrivateMessage(&event)
	}

	session.MarkMessage(msg, "")
}

func (h *pusherHandler) pushGroupMessage(event *GroupMessageEvent) {
	groupRepo := infra_repo.NewGroupRepository()
	members, err := groupRepo.FindMembersByGroupID(event.GroupID)
	if err != nil {
		log.Printf("Failed to get group members: %v", err)
		return
	}

	wsPayload := map[string]interface{}{
		"type":      "group_message",
		"msg_id":    event.MsgID,
		"group_id":  event.GroupID,
		"from_uid":  event.FromUID,
		"content":   event.Content,
		"send_time": event.SendTime,
	}

	for _, m := range members {
		if m.UID != event.FromUID {
			if redis.Cache != nil {
				online, _ := redis.Cache.IsUserOnline(m.UID)
				if !online {
					continue
				}
			}
			if globalPusher != nil {
				globalPusher.PushToUser(m.UID, wsPayload)
			}
		}
	}
}

func (h *pusherHandler) pushPrivateMessage(event *PrivateMessageEvent) {
	if redis.Cache != nil {
		online, _ := redis.Cache.IsUserOnline(event.ToUID)
		if !online {
			log.Printf("User %s is offline, skip push private message %s", event.ToUID, event.MsgID)
			return
		}
	}

	wsPayload := map[string]interface{}{
		"type":      "private_message",
		"msg_id":    event.MsgID,
		"from_uid":  event.FromUID,
		"to_uid":    event.ToUID,
		"content":   event.Content,
		"send_time": event.SendTime,
	}

	if redis.Cache != nil {
		userInfo, _ := redis.Cache.GetUserInfo(event.FromUID)
		if userInfo != nil {
			wsPayload["from_username"] = userInfo.Username
			wsPayload["from_nickname"] = userInfo.Nickname
		}
	}

	if globalPusher != nil {
		globalPusher.PushToUser(event.ToUID, wsPayload)
	}
}
