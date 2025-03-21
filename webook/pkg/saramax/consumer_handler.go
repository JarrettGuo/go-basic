package saramax

import (
	"encoding/json"
	"go-basic/webook/pkg/logger"

	"github.com/IBM/sarama"
)

type Handler[T any] struct {
	l  logger.Logger
	fn func(msg *sarama.ConsumerMessage, t T) error
}

func NewHandler[T any](l logger.Logger, fn func(msg *sarama.ConsumerMessage, t T) error) Handler[T] {
	return Handler[T]{l: l, fn: fn}
}

func (h Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}
func (h Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			h.l.Error("反序列化消息失败", logger.Error(err), logger.String("topic", msg.Topic), logger.Int32("partition", msg.Partition), logger.Int64("offset", msg.Offset))
			continue
		}
		err = h.fn(msg, t)
		// 在这里进行重试
		for i := 0; i < 3; i++ {
			err = h.fn(msg, t)
			if err == nil {
				break
			}
			h.l.Error("消费消息失败", logger.Error(err), logger.String("topic", msg.Topic), logger.Int32("partition", msg.Partition), logger.Int64("offset", msg.Offset))
		}
		if err != nil {
			h.l.Error("消费消息失败", logger.Error(err), logger.String("topic", msg.Topic), logger.Int32("partition", msg.Partition), logger.Int64("offset", msg.Offset))
		} else {
			session.MarkMessage(msg, "")
		}
	}
	return nil
}
