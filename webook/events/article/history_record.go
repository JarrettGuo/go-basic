package article

import (
	"context"
	"go-basic/webook/internal/repository"
	"go-basic/webook/pkg/logger"
	"time"

	"github.com/IBM/sarama"
)

type HistoryReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.Logger
}

func NewHistoryReadEventConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.Logger) *HistoryReadEventConsumer {
	return &HistoryReadEventConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}

func (r *HistoryReadEventConsumer) Start() error {
	panic("implement me")
}

func (r *HistoryReadEventConsumer) Consume(msg *sarama.ConsumerMessage, evt ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.repo.AddRecord(ctx, evt.Uid, evt.Aid)
}
