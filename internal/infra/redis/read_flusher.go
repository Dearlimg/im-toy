package redis

import (
	"im-chat/internal/domain/entity"
	"im-chat/internal/infra/database"
	"log"
	"time"

	"gorm.io/gorm/clause"
)

var readFlusher *ReadStatusFlusher

type ReadStatusFlusher struct {
	cache       *redisCache
	stopCh      chan struct{}
	flushPeriod time.Duration
}

func StartReadStatusFlusher(cache *redisCache) {
	readFlusher = &ReadStatusFlusher{
		cache:       cache,
		stopCh:      make(chan struct{}),
		flushPeriod: 5 * time.Second,
	}
	go readFlusher.start()
	log.Println("ReadStatusFlusher started")
}

func StopReadStatusFlusher() {
	if readFlusher != nil {
		close(readFlusher.stopCh)
		readFlusher = nil
	}
}

func (f *ReadStatusFlusher) start() {
	ticker := time.NewTicker(f.flushPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			f.flush()
		case <-f.stopCh:
			f.flush()
			return
		}
	}
}

const (
	batchSize          = 20
	maxRecordsPerFlush = 200
)

func (f *ReadStatusFlusher) flush() {
	pending, err := f.cache.PopMsgReadPending(batchSize)
	if err != nil || len(pending) == 0 {
		return
	}

	msgIDs := pending
	if len(msgIDs) == 0 {
		return
	}

	var messageReads []*entity.MessageRead
	recordCount := 0
	for _, msgID := range msgIDs {
		readers, err := f.cache.GetMsgReaders(msgID)
		if err != nil || len(readers) == 0 {
			continue
		}
		readTime := time.Now().UnixMilli()
		for _, uid := range readers {
			if recordCount >= maxRecordsPerFlush {
				break
			}
			messageReads = append(messageReads, &entity.MessageRead{
				MsgID:    msgID,
				UID:      uid,
				ReadTime: readTime,
			})
			recordCount++
		}
		if recordCount >= maxRecordsPerFlush {
			break
		}
	}

	if len(messageReads) > 0 {
		if err := database.DB.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(messageReads, 50).Error; err != nil {
			log.Printf("ReadStatusFlusher: failed to flush to DB: %v", err)
			for _, msgID := range msgIDs {
				f.cache.AddMsgReadPending(msgID, time.Now().UnixMilli())
			}
		} else {
			f.cache.RemoveMsgReadPending(msgIDs)
		}
	}
}

func AddMsgRead(msgID, uid string) error {
	if err := GetCache().AddMsgRead(msgID, uid); err != nil {
		return err
	}
	GetCache().AddMsgReadPending(msgID, time.Now().UnixMilli())
	return nil
}

func BatchGetMsgReadCount(msgIDs []string) (map[string]int64, error) {
	return GetCache().BatchGetMsgReadCount(msgIDs)
}
