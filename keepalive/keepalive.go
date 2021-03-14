package keepalive 

import (
  "net/http"
  "log"
  "fmt"
  "github.com/fatih/color"
)

func StartServer(){
  color.Cyan("Starting server on port :8080")
  http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request){
    color.Cyan("GET /")
    fmt.Fprintf(w, "Keep Alive")
  })

  log.Fatal(http.ListenAndServe(":8080", nil))
}