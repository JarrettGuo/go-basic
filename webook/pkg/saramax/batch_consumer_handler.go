package saramax

import (
	"context"
	"encoding/json"
	"go-basic/webook/pkg/logger"
	"time"

	"github.com/IBM/sarama"
)

type BatchConsumerHandler[T any] struct {
	l  logger.Logger
	fn func(msgs []*sarama.ConsumerMessage, ts []T) error
	// 用 option 模式传入
	batchSize     int
	batchDuration time.Duration
}

func NewBatchConsumerHandler[T any](l logger.Logger, fn func(msgs []*sarama.ConsumerMessage, ts []T) error) BatchConsumerHandler[T] {
	return BatchConsumerHandler[T]{
		l:  l,
		fn: fn,
		// 默认值
		batchSize:     10,
		batchDuration: time.Second,
	}
}

func (h BatchConsumerHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h BatchConsumerHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h BatchConsumerHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// 批量消费
	msgsCh := claim.Messages()
	batchSize := h.batchSize
	for {
		ctx, cancel := context.WithTimeout(context.Background(), h.batchDuration)
		done := false
		msgs := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgsCh:
				if !ok {
					cancel()
					return nil
				}
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					h.l.Error("反序列化消息失败", logger.Error(err), logger.String("topic", msg.Topic), logger.Int32("partition", msg.Partition), logger.Int64("offset", msg.Offset))
					continue
				}
				msgs = append(msgs, msg)
				ts = append(ts, t)
			}
		}
		cancel()
		if len(msgs) == 0 {
			continue
		}
		// 批量接口
		err := h.fn(msgs, ts)
		if err != nil {
			h.l.Error("消费消息失败", logger.Error(err))
		}
		for _, msg := range msgs {
			session.MarkMessage(msg, "")
		}
	}
}
