package sarama

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var address = []string{"localhost:9094"}

func TestSyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Partitioner = sarama.NewHashPartitioner
	producer, err := sarama.NewSyncProducer(address, cfg)
	assert.NoError(t, err)
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		// 主题
		Topic: "test_topic",
		// 消息的键，用于保证消息都被发送到指定分区，保证消息的顺序性
		Key: sarama.StringEncoder("test_key"),
		// 消息数据本身
		Value: sarama.StringEncoder("Hello, 这个是一条同步测试消息"),
		// 会在生产者和消费者之间传递的消息头
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("trace_key"),
				Value: []byte("trace_value"),
			},
		},
		// 只作用于发送过程，不会持久化到 Kafka 中
		Metadata: "test",
	})
	assert.NoError(t, err)
}

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	// 用于获取消息发送成功和失败的通知
	cfg.Producer.Return.Errors = true
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewAsyncProducer(address, cfg)
	require.NoError(t, err)
	msgCh := producer.Input()
	// 异步发送消息
	go func() {
		msg := &sarama.ProducerMessage{
			// 主题
			Topic: "test_topic",
			// 消息的键，用于保证消息都被发送到指定分区，保证消息的顺序性
			Key: sarama.StringEncoder("test_key"),
			// 消息数据本身
			Value: sarama.StringEncoder("Hello, 这个是一条异步测试消息"),
			// 会在生产者和消费者之间传递的消息头
			Headers: []sarama.RecordHeader{
				{
					Key:   []byte("trace_key"),
					Value: []byte("trace_value"),
				},
			},
			// 只作用于发送过程，不会持久化到 Kafka 中
			Metadata: "test",
		}
		select {
		case msgCh <- msg:
		// 防止阻塞，如果消息发送成功，不做任何处理
		default:
		}
	}()

	errCh := producer.Errors()
	succCh := producer.Successes()
	// 通过 select 语句监听消息发送的结果
	for {
		select {
		case err := <-errCh:
			t.Log("发送出了错误", err.Err)
		case msgCh := <-succCh:
			t.Log("发送成功", msgCh)
		}
	}
}

type JSONEncoder[T any] struct {
	Data T
}
