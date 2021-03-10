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
  "github.com/temoto/robotstxt"
  // "cosmic/caching"
  "encoding/json"
  "cosmic/sitemaps"
  
)


type URLQueryArray struct {
  host string
  path string
  fullURL string
}

type URLQuery struct {
  urls []URLQueryArray
  meta *Meta
}

type Meta struct {
  Name string
  Description string
  Icon string
}

type DatabaseEntryURLS struct {
  Url string
  Meta *Meta
}

type DatabaseEntry struct {
  Host string
  URLS []DatabaseEntryURLS
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

  
  // for i, entries := 0, entries; i < len(urls.urls); i++{
  //   e := DatabaseEntryURLS{
  //     url: urls.urls[i].fullURL,
  //     meta: urls.meta,
  //   }

  //   entries = append(entries, e)
  // }

  // entry, err := database.Get("host+" + u.Host)

  if err != nil{
    entry := DatabaseEntry{
      Host: u.Host,
      URLS: entries,
    }

    fmt.Println(entry)
    out, err := json.Marshal(entry)

    if err != nil{
      log.Fatal(err)
      return
    }

    fmt.Println(string(out))
    // err = database.Set("host+" + u.Host, string(out))

    // if err != nil{
    //   log.Fatal(err)
    //   return
    // }
  } else {
    fmt.Println("founnd ??")
    fmt.Println(entry)
  }
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
   
  icon = ""

  result := &Meta{
    Name: name,
    Description: description,
    Icon: icon,
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
        link := URLQueryArray{
          host: u.Host,
          path: u.Path,
          fullURL: u.Scheme + "://" + u.Host + u.Path,
        }

        urls = append(urls, link)
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