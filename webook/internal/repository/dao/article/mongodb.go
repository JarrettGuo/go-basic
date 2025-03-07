package article

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBDAO struct {
	col *mongo.Collection
	// 代表线上库
	liveCol *mongo.Collection
	// 代表雪花算法
	node *snowflake.Node
}

func NewMongoDBDAO(db *mongo.Database, node *snowflake.Node) ArticleDAO {
	return &MongoDBDAO{
		col:     db.Collection("articles"),
		liveCol: db.Collection("published_articles"),
		node:    node,
	}
}

// 在MongoDBDAO.Insert中确认
func (m *MongoDBDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now

	// 确保ID类型是int64
	if m.node == nil {
		return 0, errors.New("snowflake节点未初始化")
	}

	id := m.node.Generate().Int64()
	art.Id = id

	// 确认插入的文档结构
	doc := bson.M{
		"id":        art.Id,
		"title":     art.Title,
		"content":   art.Content,
		"author_id": art.AuthorId,
		"status":    art.Status,
		"ctime":     art.Ctime,
		"utime":     art.Utime,
	}

	_, err := m.col.InsertOne(ctx, doc)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *MongoDBDAO) UpdateById(ctx context.Context, art Article) error {
	filter := bson.M{"id": art.Id, "author_id": art.AuthorId}
	update := bson.D{bson.E{"$set", bson.M{
		"title":   art.Title,
		"content": art.Content,
		"utime":   time.Now().UnixMilli(),
		"status":  art.Status,
	}}}

	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.ModifiedCount == 0 {
		if res.MatchedCount == 0 {
			return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d", art.Id, art.AuthorId)
		}
	}
	return nil
}

func (m *MongoDBDAO) SyncStatus(ctx context.Context, id, authorId int64, status uint8) error {
	return nil
}

func (m *MongoDBDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)

	// 如果是更新操作
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		// 新建操作
		id, err = m.Insert(ctx, art)
		if err != nil {
			return 0, err
		}
	}

	// 操作线上库
	now := time.Now().UnixMilli()
	update := bson.M{
		"$set": bson.M{
			"id":        id,
			"title":     art.Title,
			"content":   art.Content,
			"author_id": art.AuthorId,
			"status":    art.Status,
			"utime":     now,
		},
		"$setOnInsert": bson.M{
			"ctime": now,
		},
	}

	filter := bson.M{"id": id}
	_, err = m.liveCol.UpdateOne(ctx, filter,
		update,
		options.Update().SetUpsert(true))

	return id, err
}

func (m *MongoDBDAO) Upsert(ctx context.Context, art PublishedArticle) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	// 构建更新操作
	update := bson.M{
		"$set": bson.M{
			"id":        art.Id,
			"title":     art.Title,
			"content":   art.Content,
			"author_id": art.AuthorId,
			"utime":     now,
			"status":    art.Status,
		},
		"$setOnInsert": bson.M{
			"ctime": now,
		},
	}
	filter := bson.M{"id": art.Id}
	// 执行upsert操作
	_, err := m.liveCol.UpdateOne(
		ctx,
		filter,
		update,
		options.Update().SetUpsert(true),
	)
	return err
}

func (m *MongoDBDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBDAO) GetById(ctx context.Context, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	index := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{Key: "author_id", Value: 1},
				bson.E{Key: "ctime", Value: 1},
			},
			Options: options.Index(),
		},
	}
	_, err := db.Collection("articles").Indexes().
		CreateMany(ctx, index)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().
		CreateMany(ctx, index)
	return err
}
