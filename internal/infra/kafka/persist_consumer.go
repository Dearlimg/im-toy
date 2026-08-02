package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"im-chat/internal/domain/entity"
	"im-chat/internal/infra/database"

	"github.com/IBM/sarama"
	"gorm.io/gorm"
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

// 批量插入配置
const (
	batchSize      = 100          // 每批最多处理100条
	batchTimeout   = 100 * time.Millisecond // 批次超时时间
)

// readBatch 批量已读事件缓存
type readBatch struct {
	events []ReadEvent
	timer  *time.Timer
	mu     sync.Mutex
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

	select {
	case <-c.handler.ready:
		log.Println("PersistConsumer started")
	case <-time.After(8 * time.Second):
		log.Println("PersistConsumer: kafka not reachable, starting without persist consumer (degraded mode)")
	}
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
		if isDuplicateKeyError(err) {
			log.Printf("Duplicate group message ignored: msg_id=%s", event.MsgID)
			return
		}
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
		if isDuplicateKeyError(err) {
			log.Printf("Duplicate private message ignored: msg_id=%s", event.MsgID)
			return
		}
		log.Printf("Failed to persist private message: %v", err)
	} else {
		log.Printf("Private message persisted: msg_id=%s", event.MsgID)
	}
}

// 全局批量处理器
var (
	batchProcessor *readBatchProcessor
)

type readBatchProcessor struct {
	batches map[string]*readBatch // key: uid
	mu      sync.RWMutex
}

func init() {
	batchProcessor = &readBatchProcessor{
		batches: make(map[string]*readBatch),
	}
}

// addToBatch 将已读事件添加到批次中
func (p *readBatchProcessor) addToBatch(event *ReadEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()

	batch, exists := p.batches[event.UID]
	if !exists {
		batch = &readBatch{
			events: make([]ReadEvent, 0, batchSize),
		}
		p.batches[event.UID] = batch
	}

	batch.mu.Lock()
	batch.events = append(batch.events, *event)

	// 如果达到批次大小或还没有设置定时器，则触发批量插入
	if len(batch.events) >= batchSize || batch.timer == nil {
		if batch.timer != nil {
			batch.timer.Stop()
		}
		batch.timer = time.AfterFunc(batchTimeout, func() {
			p.flushBatch(event.UID)
		})
	}
	batch.mu.Unlock()
}

// flushBatch 刷新批次到数据库
func (p *readBatchProcessor) flushBatch(uid string) {
	p.mu.Lock()
	batch, exists := p.batches[uid]
	if !exists {
		p.mu.Unlock()
		return
	}
	delete(p.batches, uid)
	p.mu.Unlock()

	batch.mu.Lock()
	events := make([]ReadEvent, len(batch.events))
	copy(events, batch.events)
	batch.events = batch.events[:0]
	if batch.timer != nil {
		batch.timer.Stop()
		batch.timer = nil
	}
	batch.mu.Unlock()

	if len(events) == 0 {
		return
	}

	p.persistBatchReadEvents(events)
}

// persistBatchReadEvents 批量插入已读记录
func (p *readBatchProcessor) persistBatchReadEvents(events []ReadEvent) {
	if len(events) == 0 {
		return
	}

	// 1. 批量查询已存在的记录
	msgIDs := make([]string, len(events))
	uid := events[0].UID
	for i, event := range events {
		msgIDs[i] = event.MsgID
	}

	var existingReads []entity.MessageRead
	if err := database.DB.Where("uid = ? AND msg_id IN ?", uid, msgIDs).Find(&existingReads).Error; err != nil {
		log.Printf("Failed to query existing read events: %v", err)
		return
	}

	// 2. 构建已存在的 msg_id 集合
	existingSet := make(map[string]bool)
	for _, r := range existingReads {
		existingSet[r.MsgID] = true
	}

	// 3. 过滤出需要插入的记录
	var toInsert []entity.MessageRead
	for _, event := range events {
		if !existingSet[event.MsgID] {
			toInsert = append(toInsert, entity.MessageRead{
				MsgID:    event.MsgID,
				UID:      event.UID,
				ReadTime: event.ReadTime,
			})
		}
	}

	// 4. 批量插入（使用事务提高性能）
	if len(toInsert) > 0 {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			return tx.Create(&toInsert).Error
		})

		if err != nil {
			// 检查是否是重复键错误
			if isDuplicateKeyError(err) {
				log.Printf("Some read events already exist, skipped")
				return
			}
			log.Printf("Failed to batch persist read events (count=%d): %v", len(toInsert), err)
			return
		}
		log.Printf("Batch persisted %d read events (skipped %d existing)", len(toInsert), len(existingSet))
	}
}

func (h *persistHandler) persistReadEvent(event *ReadEvent) {
	batchProcessor.addToBatch(event)
}

func isDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "Duplicate entry")
}
