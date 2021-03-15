package main

import (
  "fmt"
  "time"
  "log"
  "cosmic/database"
  "cosmic/keepalive"
  "cosmic/scraping"
  "os"

  "go.mongodb.org/mongo-driver/mongo"
  "context"
)



func main() {
  ascii := `

                                             /$$          
                                            |__/          
  /$$$$$$$  /$$$$$$   /$$$$$$$ /$$$$$$/$$$$  /$$  /$$$$$$$
 /$$_____/ /$$__  $$ /$$_____/| $$_  $$_  $$| $$ /$$_____/
| $$      | $$  \ $$|  $$$$$$ | $$ \ $$ \ $$| $$| $$      
| $$      | $$  | $$ \____  $$| $$ | $$ | $$| $$| $$      
|  $$$$$$$|  $$$$$$/ /$$$$$$$/| $$ | $$ | $$| $$|  $$$$$$$
 \_______/ \______/ |_______/ |__/ |__/ |__/|__/ \_______/
                                                          
                                                          
                                                          

  `

  fmt.Println(ascii)

  isTests := os.Getenv("TEST") == "yes"

  var Client *mongo.Client
  var ctx context.Context

  if isTests == false{
    var err error
    Client, ctx, err = database.Connect()

    if err != nil{
      log.Fatal(err)
    } 
  }
  
  ticker := time.NewTicker(5 * time.Second)
  quit := make(chan struct{})

  go func() {
      for {
        select {
          case <- ticker.C:
            go scraping.PrintMemUsage()
            
          case <- quit:
              ticker.Stop()
              return
          }
      }
  }()

  if isTests == false{
    go scraping.Execute("https://repl.it", Client, ctx)
    go scraping.Execute("https://google.com", Client, ctx)
    go scraping.Execute("https://github.com", Client, ctx)
    go scraping.Execute("https://myflixer.to", Client, ctx) 
  } else {
    go scraping.Execute("https://example.com", nil, nil)
  }

  if isTests == false{
    go scraping.ReScrape(Client, ctx)

    ticker = time.NewTicker(5 * time.Second)
    quit = make(chan struct{})

    go func() {
        for {
          select {
            case <- ticker.C:
              go scraping.ReScrape(Client, ctx)
              
            case <- quit:
                ticker.Stop()
                return
            }
        }
    }()
  }

  keepalive.StartServer(isTests)
  

}