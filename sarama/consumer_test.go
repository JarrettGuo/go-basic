package sarama

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"log"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	// 新建消费者
	consumer, err := sarama.NewConsumerGroup(address, "test_consumer", cfg)
	require.NoError(t, err)

	// 带超时的 context
	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Minute*10, func() {
		cancel()
	})
	// 开始消费消息
	err = consumer.Consume(ctx, []string{"test_topic"}, testConsumerGroupHandler{})
	// 消费结束就会到这里
	t.Log(err, time.Since(start).String())
}

type testConsumerGroupHandler struct {
}

func (t testConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	partitions := session.Claims()["test_topic"]
	for _, partition := range partitions {
		// 从指定分区的指定位置开始消费
		session.ResetOffset("test_topic", partition, sarama.OffsetOldest, "")
	}
	return nil
}

func (t testConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("cleanup")
	return nil
}

// 异步消费
func (t testConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		// 这里是为了解决闭包问题，否则会导致 msg 一直是最后一条消息
		msg := msg
		go func() {
			// 消费msg
			log.Println(string(msg.Value))
			session.MarkMessage(msg, "")
		}()
	}

	const batchSize = 10
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		var eg errgroup.Group
		var last *sarama.ConsumerMessage
		done := false
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				// 这边代表超时
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					// msgs 被关闭，也就是退出消费逻辑
					return nil
				}
				last = msg
				eg.Go(func() error {
					time.Sleep(time.Second)
					log.Println(string(msg.Value))
					return nil
				})
			}
		}
		err := eg.Wait()
		if err != nil {
			// 记录日志
			continue
		}
		if last != nil {
			session.MarkMessage(last, "")
		}
	}
	// msgs 被关闭，也就是退出消费逻辑
	return nil
}

// 单个消费者的消费逻辑
// session 代表和 Kafka 服务器的连接的那段时间   claim 代表一次消费的请求
func (t testConsumerGroupHandler) ConsumeClaimV1(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		var bizMsg MyBizMsg
		err := json.Unmarshal(msg.Value, &bizMsg)
		if err != nil {
			// 消费消息出错，做不了什么，大多数情况下只能重试，最后记录日志
			continue
		}
		// 将消息交给业务处理
		session.MarkMessage(msg, "")
	}
	// msgs 被关闭，也就是退出消费逻辑
	return nil
}

type MyBizMsg struct {
	Name string
}
