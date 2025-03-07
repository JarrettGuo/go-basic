package article

import (
	"context"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReaderDAO interface {
	Upsert(ctx context.Context, art PublishedArticle) error
}

func NewReaderDAO(mdb *mongo.Database, node *snowflake.Node) ReaderDAO {
	return &MongoDBDAO{
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
		node:    node,
	}
}
