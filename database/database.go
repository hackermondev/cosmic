package database

import (
  "context"
	"time"
  "os"

  "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Connect() (*mongo.Client, context.Context, error){
  client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB")))

	if err != nil {
		return client, nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	err = client.Connect(ctx)

	if err != nil {
		return client, nil, err
	}

	// defer client.Disconnect(ctx)

  err = client.Ping(ctx, readpref.Primary())

  if err != nil {
    return client, nil, err
  }


  return client, ctx, nil
}

func AddToCollection(collectionName string, item interface{}) (bool, error) {
  client, ctx, err := Connect()

  if err != nil{
    return false, err
  }

  collection := client.Database("cosmic").Collection(collectionName)

  _, err = collection.InsertOne(ctx, item)

  if err != nil{
    return false, err
  }

  return true, nil
}

func GetFromCollection(collectionName string, filter bson.D)  (*mongo.Cursor, context.Context, error){
  client, ctx, err := Connect()

  if err != nil{
    return nil, nil, err
  }

  collection := client.Database("cosmic").Collection(collectionName)

  cur, err := collection.Find(ctx, filter)

  if err != nil{
    return nil, nil, err
  }

  return cur, ctx, nil
}