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

	// monitor := &event.CommandMonitor{
	// 	Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
	// 		fmt.Println(evt.Command)
	// 	},
	// }
	opts := options.Client().
		ApplyURI("mongodb://root:root@localhost:27017/")
		// SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	return client.Database("webook")
}

func InitSnowflakeNode() *snowflake.Node {
	// 通常使用一个确定的节点ID，如服务器ID或进程ID
	// 这里使用1作为示例
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	return node
}
