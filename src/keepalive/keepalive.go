package keepalive 

import (
  "net/http"
  "log"
  "fmt"
  "github.com/fatih/color"
  "net/url"
  "encoding/json"
  "io/ioutil"
  "cosmic/scraping"
  "cosmic/database"
  "os"
)

type Request struct {
  url string
}

func StartServer(){
  port := os.Getenv("PORT")

  if port == ""{
    port = "3868"
  }

  color.Cyan("Starting server on port :" + port)

  Client, ctx, err := database.Connect()
  
  if err != nil{
    log.Fatal(err)
  }

  http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request){
    color.Cyan("GET /")
    fmt.Fprintf(w, "Keep Alive")
  })

  http.HandleFunc("/scrape", func (w http.ResponseWriter, r *http.Request){
    if err := r.ParseForm(); err != nil{
      fmt.Fprintf(w, "Error while parsing form")
      return
    }

    u := r.FormValue("url")

    if u == ""{
      body, err := ioutil.ReadAll(r.Body)

      if err != nil {
        fmt.Fprintf(w, "Error while parsing form")
        return
      }

      var ra Request

      if err := json.Unmarshal(body, &ra); err != nil{
        fmt.Fprintf(w, "Error while parsing form")
        return
      }

      fmt.Println(ra)
      u = ra.url
    }

    
    parse, err := url.Parse(u)    

    if err != nil{
      fmt.Fprintf(w, "Error while parsing form")
      return
    }


    fullurl := parse.Scheme + "://" + parse.Host + parse.Path

    fmt.Println("Scraping " + fullurl + " , requested by " + r.RemoteAddr)

    go scraping.Execute(fullurl, Client, ctx)

    fmt.Fprintf(w, ":)")
  })

  log.Fatal(http.ListenAndServe(":" + port, nil))
}