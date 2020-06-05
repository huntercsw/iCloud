package commons

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"iCloud/conf"
	"time"
)

type MONGO struct {
	cli *mongo.Client
}

func (m *MONGO) MongoInit() (err error) {
	var (
		ctx context.Context
	)
	uri := "mongodb://" + conf.Iconf.Mongo
	ctx, _ = context.WithTimeout(context.TODO(), 5*time.Second)
	if m.cli, err = mongo.Connect(ctx, options.Client().ApplyURI(uri).SetMaxPoolSize(5).SetMinPoolSize(2)); err != nil {
		return
	}
	return
}