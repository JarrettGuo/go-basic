package article

import (
	"context"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthorDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
}

func NewAuthorDAO(mdb *mongo.Database, node *snowflake.Node) AuthorDAO {
	return &MongoDBDAO{
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
		node:    node,
	}
}
