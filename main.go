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
  // "go.mongodb.org/mongo-driver/bson"
  "github.com/temoto/robotstxt"
  // "cosmic/caching"
  // "encoding/json"
  "cosmic/sitemaps"
  "cosmic/database"
  "cosmic/keepalive"
  "math" 

  "strconv"
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
  Node int
}

func min(a, b int) int {
    if a <= b {
        return a
    }
    return b
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
  var toBeScraped []string

  if urls == nil{
    return
  }

  for k := range urls.urls {
    p, _ := url.Parse(urls.urls[k].fullURL)

    toBeScraped = append(toBeScraped, urls.urls[k].fullURL)

    if p.Host != u.Host{

    } else {
      e := DatabaseEntryURLS{
        Url: urls.urls[k].fullURL,
        Meta: urls.urls[k].Meta,
      }

      entries = append(entries, e)
    }
  }

  
  // for i, entries := 0, entries; i < len(urls.urls); i++{
  //   e := DatabaseEntryURLS{
  //     Url: urls.urls[i].fullURL,
  //     Meta: urls.urls[i].Meta,
  //   }

  //   entries = append(entries, e)
  // }

  var entryNodes []DatabaseEntry

  if len(entries) < 500{
    entry := DatabaseEntry{
      Host: u.Host,
      URLS: entries,
      Meta: urls.meta,
      Node: 0,
    }

    entryNodes = append(entryNodes, entry)
  } else {
    times := math.Round(float64(len(entries))/500)

    for i := 0; i <= int(times); i++{
      entry := DatabaseEntry{
        Host: u.Host,
        URLS: entries[i:min(i+500, len(entries))],
        Meta: urls.meta,
        Node: i,
      }

      entryNodes = append(entryNodes, entry)
    }
  }

  // entry := DatabaseEntry{
  //   Host: u.Host,
  //   URLS: entries,
  //   Meta: urls.meta,
  // }

  // fmt.Println(entry)

  if err != nil{
    log.Fatal(err)
    return
  }

  database.DeleteIfExists("hosts", u.Host)

  for a := range entryNodes{
    _, err := database.AddToCollection("hosts", entryNodes[a])

    if err != nil{
      log.Fatal(err)
    }
  }

  // _, err = database.AddToCollectionAndDeleteIfExists("hosts", entryNodes, u.Host)

  // if err != nil {
  //   log.Fatal(err)
  // }

  // fmt.Println("Saving current data")
  for x := range toBeScraped{
    Execute(toBeScraped[x])
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
  
  if err != nil{
    return "", err
  }

  if resp.StatusCode != 200{

    if resp.StatusCode == 429{
      retryAfter := resp.Header.Get("Retry-After")

      retry, err := strconv.Atoi(retryAfter)

      if err != nil{

      } else {
        time.Sleep(time.Duration(retry * 1000))

        s, err := getRobotsTxt(rawurl)

        return s, err
      }
    }

    return "", errors.New("Website returns status code: " + resp.Status + " ( " + rawurl + ")")
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

  if err != nil{
    return "", err
  }

  if resp.StatusCode != 200{
    if resp.StatusCode != 200{

    if resp.StatusCode == 429{
      retryAfter := resp.Header.Get("Retry-After")

      retry, err := strconv.Atoi(retryAfter)

      if err != nil{

      } else {
        time.Sleep(time.Duration(retry * 1000))

        s, err := Request(rawurl)

        return s, err
      }
    }

    return "", errors.New("Website returns status code: " + resp.Status + " ( " + rawurl + ")")
  }

    return "", errors.New("Website returns status code: " + resp.Status + " ( " + rawurl + ")")
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)

  if err != nil{
    return "", err
  }

  return string(body), nil
}

func scrapeURL(rawurl string) (*URLQuery, error){
  // fmt.Println(rawurl)
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
  
  if err != nil{
    return nil, err
  }

  if resp.StatusCode != 200{
    if resp.StatusCode != 200{

    if resp.StatusCode == 429{
      retryAfter := resp.Header.Get("Retry-After")

      retry, err := strconv.Atoi(retryAfter)

      if err != nil{

      } else {
        time.Sleep(time.Duration(retry * 1000))

        s, err := scrapeURL(rawurl)

        return s, err
      }
    }

    return nil, errors.New("Website returns status code: " + resp.Status + " ( " + rawurl + ")")
  }

    return nil, errors.New("Website returns status code: " + resp.Status + " ( " + rawurl + ")")
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
      // fmt.Println("Could not parse sitemap ", robots.Sitemaps[sitemap])

      panic("ok")
    } else {
      for u := range su{
      link := su[u]
      
      u, err := url.Parse(link)


      if err != nil{
        
      } else {
        b, err := Request(u.Scheme + "://" + u.Host + u.Path)

        if err != nil{
          fmt.Println("Could not get meta for ", u.Host + u.Path)

          fmt.Println(err)

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
  }
  // fmt.Println(robots.Sitemaps)

  
  group := robots.FindGroup("cosmicbot")

  rawurlparse, err := url.Parse(rawurl)

  if err != nil{
    return nil, err
  }

  var result *URLQuery
  
  if !group.Test(rawurlparse.Path){
    return result, nil
  }

  doc.Find("a").Each(func (i int, s *goquery.Selection){
    link, _ := s.Attr("href")

    if strings.HasPrefix(link, "/") == true && strings.HasPrefix(link, "//") == false{
      u, err := url.Parse(rawurl)


      if err != nil{
        return
      }

      if !group.Test(link){
        return
      }

      // link = url + link

      b, err := Request(u.Scheme + "://" + u.Host + link)
      
      if err != nil{
        fmt.Println(err)

        b = "<html><b>invalid</b></html>"
      }

      meta, err := getMeta(b)

      if err != nil{
        log.Fatal(err)
      }

      link := URLQueryArray{
        host: u.Host,
        path: link,
        fullURL: u.Scheme + "://" + u.Host + link,
        Meta: meta,
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


      b, err := Request(u.Scheme + "://" + u.Host + u.Path)

      if err != nil{
        fmt.Println(err)
      }

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
  })  

  result = &URLQuery{
    urls: urls,
    meta: meta,
  }

  return result, nil
}

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

  database.Connect()

  go Execute("https://repl.it")
  go Execute("https://google.com")
  go Execute("https://github.com")
  go Execute("https://myflixer.to")
  
  keepalive.StartServer()
}