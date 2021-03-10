package database

import (
  "context"
	"fmt"
	"log"
	"time"
  "os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Connect() (mongo.Database,error){
  client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB")))

	if err != nil {
		return nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err = client.Connect(ctx)

	if err != nil {
		return nil, err
	}

	defer client.Disconnect(ctx)

  err = client.Ping(ctx, readpref.Primary())

  if err != nil {
    return nil, err
  }

  db := client.Database("cosmic-database")


  return db
}