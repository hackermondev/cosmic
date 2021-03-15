package main

import (
  "fmt"
  "time"
  "log"
  "cosmic/database"
  "cosmic/keepalive"
  "cosmic/scraping"
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

  Client, ctx, err := database.Connect()

  if err != nil{
    log.Fatal(err)
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

  go scraping.Execute("https://repl.it", Client, ctx)
  go scraping.Execute("https://google.com", Client, ctx)
  go scraping.Execute("https://github.com", Client, ctx)
  go scraping.Execute("https://myflixer.to", Client, ctx) 

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

  keepalive.StartServer()
  
}