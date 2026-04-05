package kafka

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/IBM/sarama"
	"im-chat/internal/domain/entity"
	"im-chat/internal/infra/database"
)

type PersistConsumer struct {
	reader  sarama.ConsumerGroup
	topics  []string
	handler *persistHandler
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

type persistHandler struct {
	ready chan bool
}

func NewPersistConsumer(brokers []string, groupID string, topics []string) (*PersistConsumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	reader, err := sarama.NewConsumerGroup(brokers, groupID, saramaConfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &PersistConsumer{
		reader:  reader,
		topics:  topics,
		handler: &persistHandler{ready: make(chan bool)},
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (c *PersistConsumer) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.reader.Consume(c.ctx, c.topics, c.handler); err != nil {
				log.Printf("PersistConsumer error: %v", err)
			}
			if c.ctx.Err() != nil {
				return
			}
			c.handler.ready = make(chan bool)
		}
	}()

	<-c.handler.ready
	log.Println("PersistConsumer started")
}

func (c *PersistConsumer) Stop() error {
	c.cancel()
	c.wg.Wait()
	return c.reader.Close()
}

func (h *persistHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *persistHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *persistHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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

func (h *persistHandler) processMessage(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	switch msg.Topic {
	case TopicGroupMsgRaw:
		var event GroupMessageEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal group message: %v", err)
			return
		}
		h.persistGroupMessage(&event)

	case TopicPrivateMsg:
		var event PrivateMessageEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal private message: %v", err)
			return
		}
		h.persistPrivateMessage(&event)

	case TopicReadEvents:
		var event ReadEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Failed to unmarshal read event: %v", err)
			return
		}
		h.persistReadEvent(&event)
	}

	session.MarkMessage(msg, "")
}

func (h *persistHandler) persistGroupMessage(event *GroupMessageEvent) {
	msg := &entity.GroupMessage{
		MsgID:    event.MsgID,
		GroupID:  event.GroupID,
		FromUID:  event.FromUID,
		Content:  event.Content,
		Type:     2,
		SendTime: event.SendTime,
	}

	if err := database.DB.Create(msg).Error; err != nil {
		log.Printf("Failed to persist group message: %v", err)
	} else {
		log.Printf("Group message persisted: msg_id=%s, group_id=%s", event.MsgID, event.GroupID)
	}
}

func (h *persistHandler) persistPrivateMessage(event *PrivateMessageEvent) {
	msg := &entity.PrivateMessage{
		MsgID:    event.MsgID,
		FromUID:  event.FromUID,
		ToUID:    event.ToUID,
		Content:  event.Content,
		Type:     1,
		SendTime: event.SendTime,
		Read:     0,
	}

	if err := database.DB.Create(msg).Error; err != nil {
		log.Printf("Failed to persist private message: %v", err)
	} else {
		log.Printf("Private message persisted: msg_id=%s", event.MsgID)
	}
}

func (h *persistHandler) persistReadEvent(event *ReadEvent) {
	read := &entity.MessageRead{
		MsgID:    event.MsgID,
		UID:      event.UID,
		ReadTime: event.ReadTime,
	}

	if err := database.DB.Create(read).Error; err != nil {
		log.Printf("Failed to persist read event: %v", err)
	}
}
