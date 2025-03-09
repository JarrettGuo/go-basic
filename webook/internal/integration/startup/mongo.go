package startup

import (
	"context"
	"time"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB() *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Client().
		ApplyURI("mongodb://root:root@localhost:27017/")
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	return client.Database("webook")
}

func InitSnowflakeNode() *snowflake.Node {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	return node
}
