package main

import (
  "fmt"
  "net/http"
  "time"
  "errors"
  goquery "github.com/PuerkitoBio/goquery"
  "strings"
  "log"
  "net/url"
  s "github.com/hackermondev/cosmic/structs"
)


func scrapeURL(rawurl string) ([]s.URLQuery, error){
  request, err := http.NewRequest("GET", rawurl, nil)

  client := &http.Client{
    Timeout: 30 * time.Second,
  }

  if err != nil{
    return nil, err
  }

  request.Header.Set("User-Agent", "cosmicbot")
  request.Header.Set("referer", rawurl)
  
  resp, err := client.Do(request)

  if resp.StatusCode != 200{
    return nil, errors.New("Website returns status code: ")
  }

  doc, err := goquery.NewDocumentFromReader(resp.Body)

  if err != nil{
    return nil, err
  }

  var urls []s.URLQuery

  doc.Find("a").Each(func (i int, s *goquery.Selection){
    link, _ := s.Attr("href")

    if strings.HasPrefix(link, "/") {
      u, err := url.Parse(rawurl)


      if err != nil{
        return
      }

      // link = url + link

      link := s.URLQuery{
        host: u.Host,
        path: link,
        fullURL: rawurl + link,
      }

      urls = append(urls, link)
    } else {
      u, err := url.Parse(link)

      if err != nil{
        return
      }

      if u.Scheme != "http" && u.Scheme != "https"{
        return
      }

      link := s.URLQuery{
        host: u.Host,
        path: link,
        fullURL: link,
      }

      urls = append(urls, link)
    }
  })  

  return urls, nil
}

func main() {
  urls, err := scrapeURL("https://google.com")

  if err != nil{
    log.Fatal(err)
    return
  }


  fmt.Println(urls)
}