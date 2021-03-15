package database

import (
  "context"
	"time"
  "os"

  "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

  "github.com/fatih/color"
)


func Connect() (*mongo.Client, context.Context, error){
  client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB")))

	if err != nil {
		return client, nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Minute)

	err = client.Connect(ctx)

	if err != nil {
		return client, nil, err
	}

	// defer client.Disconnect(ctx)

  err = client.Ping(ctx, readpref.Primary())

  if err != nil {
    return client, nil, err
  }

  color.Blue("Database is connected to the server.")
  return client, ctx, nil
}


func AddToCollection(client *mongo.Client, ctx context.Context, collectionName string, item interface{}) (bool, error) {

  // client, ctx, err := Connect()

  // if err != nil{
  //   return false, err
  // }

  collection := client.Database("cosmic").Collection(collectionName)

  _, err := collection.InsertOne(ctx, item)

  if err != nil{
    return false, err
  }

  return true, nil
}

func DeleteIfExists(client *mongo.Client, ctx context.Context,collectionName string, host string) (bool, error){
  // client, ctx, err := Connect()

  // if err != nil{
  //   return false, err
  // }

  collection := client.Database("cosmic").Collection(collectionName)

  cur, err := collection.Find(ctx, bson.M{ "host":  host })

  if err != nil{
    cur = nil
  }

  if cur != nil{
    _, err := collection.DeleteMany(ctx, bson.M{"host": host })
    
    // client.Disconnect(ctx)
    if err != nil{
      return false, err
    }
  }

  // client.Disconnect(ctx)
  return true, nil
}

func GetFromCollection(client *mongo.Client, ctx context.Context, collectionName string, filter bson.D)  (*mongo.Cursor, context.Context, error){
  // client, ctx, err := Connect()

  // if err != nil{
  //   return nil, nil, err
  // }

  collection := client.Database("cosmic").Collection(collectionName)

  cur, err := collection.Find(ctx, filter)
  
  // client.Disconnect(ctx)
  if err != nil{
    return nil, nil, err
  }

  return cur, ctx, nil
}