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
  "io/ioutil"
  "go.mongodb.org/mongo-driver/bson"
  "github.com/temoto/robotstxt"
  // "cosmic/caching"
  // "encoding/json"
  "cosmic/sitemaps"
  "cosmic/database" 
)


type URLQueryArray struct {
  host string
  path string
  fullURL string
  Meta *Meta
}

type URLQuery struct {
  urls []URLQueryArray
  meta *Meta
}

type Meta struct {
  Name string
  Description string
  Icon string
  Keywords []string
}

type DatabaseEntryURLS struct {
  Url string
  Meta *Meta
}

type DatabaseEntry struct {
  Host string
  URLS []DatabaseEntryURLS
  Meta *Meta
}

func Execute(rawurl string){
  u, err := url.Parse(rawurl)

  if err != nil{
    log.Fatal(err)
    return
  }

  urls, err := scrapeURL(rawurl)

  if err != nil{
    log.Fatal(err)
    return
  }

  var entries []DatabaseEntryURLS

  for k := range urls.urls {
    e := DatabaseEntryURLS{
      Url: urls.urls[k].fullURL,
      Meta: urls.meta,
    }

    entries = append(entries, e)
  }

  
  for i, entries := 0, entries; i < len(urls.urls); i++{
    e := DatabaseEntryURLS{
      Url: urls.urls[i].fullURL,
      Meta: urls.urls[i].Meta,
    }

    entries = append(entries, e)
  }

  cur, ctx, err := database.GetFromCollection("hosts", bson.D{{}})

  if err != nil{
    log.Fatal(err)
  }

  // var entriesa []DatabaseEntry

  if err := cur.Err(); err != nil{
    log.Fatal(err)
  }

  for cur.Next(ctx) {
    var entry DatabaseEntry
    err = cur.Decode(&entry)

    if err != nil{
      log.Fatal(err)
      return
    }

    fmt.Println(entry)
  }

  cur.Close(ctx)

  entry := DatabaseEntry{
    Host: u.Host,
    URLS: entries,
    Meta: urls.meta,
  }

  fmt.Println(entry)
  // out, err := json.Marshal(entry)

  // if err != nil{
  //   log.Fatal(err)
  //   return
  // }

  // b, err := database.AddToCollection("hosts", entry)

  // if err != nil{
  //   log.Fatal(err)
  // }

  // fmt.Println(b)
}

func getMeta(source string) (*Meta, error){
  var name string
  var description string
  var icon string

  doc, err := goquery.NewDocumentFromReader(strings.NewReader(source))

  if err != nil{
    return nil, err
  }  

  name = doc.Find("title").Text()

  description, _ = doc.Find("meta[name='description']").Attr("content")
  
  keywords, _ := doc.Find("meta[name='keywords']").Attr("content")

  icon, e := doc.Find("link[rel='icon']").Attr("href")

  if e == false{
    icon, _ =  doc.Find("link[rel='shortcut icon']").Attr("href")
  }

  result := &Meta{
    Name: name,
    Description: description,
    Icon: icon,
    Keywords: strings.Split(keywords, ","),
  }

  return result, nil
}

func getRobotsTxt(rawurl string) (string ,error){
  u, err := url.Parse(rawurl)


  if err != nil{
    return "", err
  }

  rawurl = u.Scheme + "://" + u.Host + "/robots.txt"

  request, err := http.NewRequest("GET", rawurl, nil)

  client := &http.Client{
    Timeout: 30 * time.Second,
  }

  if err != nil{
    return "", err
  }

  request.Header.Set("User-Agent", "cosmicbot (+github.com/hackermondev/cosmic)")
  request.Header.Set("referer", rawurl)
  
  resp, err := client.Do(request)

  if resp.StatusCode != 200{
    return "", errors.New("Website returns status code: ")
  }

  text, _ := ioutil.ReadAll(resp.Body)

  return string(text), nil
}

func Request(rawurl string) (string, error){
  request, err := http.NewRequest("GET", rawurl, nil)

  client := &http.Client{
    Timeout: 30 * time.Second,
  }

  if err != nil{
    return "", err
  }

  request.Header.Set("User-Agent", "cosmicbot (+github.com/hackermondev/cosmic)")
  request.Header.Set("referer", rawurl)
  
  resp, err := client.Do(request)

  if resp.StatusCode != 200{
    return "", errors.New("Website returns status code: ")
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)

  if err != nil{
    return "", err
  }

  return string(body), nil
}

func scrapeURL(rawurl string) (*URLQuery, error){
  request, err := http.NewRequest("GET", rawurl, nil)

  client := &http.Client{
    Timeout: 30 * time.Second,
  }

  if err != nil{
    return nil, err
  }

  request.Header.Set("User-Agent", "cosmicbot (+github.com/hackermondev/cosmic)")
  request.Header.Set("referer", rawurl)
  
  resp, err := client.Do(request)

  if resp.StatusCode != 200{
    return nil, errors.New("Website returns status code: ")
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)

  if err != nil{
    return nil, err
  }

  meta, err := getMeta(string(body))

  //fmt.Println(string(body))
  if err != nil{
    return nil, err
  }

  doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))

  if err != nil{
    return nil, err
  }

  var urls []URLQueryArray

  rtxt, _ := getRobotsTxt(rawurl)

  robots, err := robotstxt.FromString(rtxt)

  if err != nil{
    return nil, err
  }

  for sitemap := range robots.Sitemaps{
    su, err := sitemaps.GetURLS(robots.Sitemaps[sitemap])

    if err != nil{
      log.Fatal(err)
      return nil, err
    }

    for u := range su{
      link := su[u]
      
      u, err := url.Parse(link)


      if err != nil{
        
      } else {
        b, err := Request(u.Scheme + "://" + u.Host + u.Path)

        if err != nil{
          fmt.Println("Could not get meta for ", u.Host + u.Path)

          var meta *Meta
          link := URLQueryArray{
            host: u.Host,
            path: u.Path,
            fullURL: u.Scheme + "://" + u.Host + u.Path,
            Meta: meta,
          }

          urls = append(urls, link)
        } else {
          meta, err := getMeta(b)
          
          if err != nil{
            log.Fatal(err)
          }

          link := URLQueryArray{
            host: u.Host,
            path: u.Path,
            fullURL: u.Scheme + "://" + u.Host + u.Path,
            Meta: meta,
          }

          urls = append(urls, link)
        }
      }

    }
  }
  // fmt.Println(robots.Sitemaps)

  
  group := robots.FindGroup("cosmicbot")

  rawurlparse, err := url.Parse(rawurl)

  if err != nil{
    return nil, err
  }

  if !group.Test(rawurlparse.Path){
    return nil, errors.New("Cannot read this page")
  }

  doc.Find("a").Each(func (i int, s *goquery.Selection){
    link, _ := s.Attr("href")

    if strings.HasPrefix(link, "/") {
      u, err := url.Parse(rawurl)


      if err != nil{
        return
      }

      if !group.Test(link){
        return
      }

      // link = url + link

      link := URLQueryArray{
        host: u.Host,
        path: link,
        fullURL: u.Scheme + "://" + u.Host + link,
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

      rawurlparse, err := url.Parse(rawurl)

      if err != nil{
        return
      }

      if u.Host == rawurlparse.Host{
        group = group

        if !group.Test(link){
          return
        }
      } else {
        rtxt, _ := getRobotsTxt(link)

        robots, err := robotstxt.FromString(rtxt)

        if err != nil{
          return
        }

        group := robots.FindGroup("cosmicbot")

        if !group.Test(link){
          return
        } 
      }


      link := URLQueryArray{
        host: u.Host,
        path: u.Path,
        fullURL: u.Scheme + "://" + u.Host + u.Path,
      }

      urls = append(urls, link)
    }
  })  

  result := &URLQuery{
    urls: urls,
    meta: meta,
  }

  return result, nil
}

func main() {
  database.Connect()

  Execute("https://replit.com/")
}