package caching 

import (
  "github.com/go-redis/redis/v8"
  "os"
  "context"
  "time"
  // "fmt"
)

var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
  Addr: os.Getenv("REDIS_DATABASE_HOST"),
  Password: os.Getenv("REDIS_DATABASE_PASSWORD"),
  DB: 0,
})

func Connect() *redis.Client{
  rdb := redis.NewClient(&redis.Options{
    Addr: os.Getenv("REDIS_DATABASE_HOST"),
    Password: os.Getenv("REDIS_DATABASE_PASSWORD"),
    DB: 0,
  })

  return rdb
}


func Set(name string, value string, expiresAt time.Duration) error{
  // rdb := Connect()


  err := rdb.Set(ctx, name, value, expiresAt).Err()

  if err != nil {
    return err
  }

  return nil
}

func Get(name string) (string, error){
  // rdb := Connect()
  val, err := rdb.Get(ctx, name).Result()

  if err != nil {
    return "", err
  }

  return val, nil
}