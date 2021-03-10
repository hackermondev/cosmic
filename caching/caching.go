package caching 

import (
  "github.com/go-redis/redis/v8"
  "os"
  "context"
)

var ctx = context.Background()

func Connect() *redis.Client{
  rdb := redis.NewClient(&redis.Options{
    Addr: os.Getenv("DATABASE_HOST"),
    Password: os.Getenv("DATABASE_PASSWORD"),
    DB: 0,
  })

  return rdb
}


func Set(name string, value string) error{
  rdb := Connect()
  err := rdb.Set(ctx, "key", "value", 0).Err()

  if err != nil {
    return err
  }

  return nil
}

func Get(name string) (string, error){
  rdb := Connect()
  val, err := rdb.Get(ctx, name).Result()

  if err != nil {
    return "", err
  }

  return val, nil
}