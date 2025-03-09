package article

import (
	"context"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type ReaderDAO interface {
	Upsert(ctx context.Context, art PublishedArticle) error
}

func NewReaderDAOV1(mdb *mongo.Database, node *snowflake.Node) ReaderDAO {
	return &MongoDBDAO{
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
		node:    node,
	}
}

func NewReaderDAO(db *gorm.DB) ReaderDAO {
	return &GORMArticleDAO{
		db: db,
	}
}
