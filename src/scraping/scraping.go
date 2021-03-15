package scraping

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
  "cosmic/caching"
  // "encoding/json"
  "cosmic/sitemaps"
  "cosmic/database"
  "math" 

  "strconv"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "context"

  "runtime"
  "os"
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
  Language string
  Secure bool
  Score int
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


func bToMb(b uint64) uint64 {
  return b / 1024 / 1024
}

func PrintMemUsage() {
  var m runtime.MemStats

  runtime.ReadMemStats(&m)

  fmt.Println("--------------------------------")
  fmt.Println("Alloc = %v MiB", bToMb(m.Alloc))
  fmt.Println("TotalAlloc = %v MiB", bToMb(m.TotalAlloc))
  fmt.Println("Sys = %v MiB", bToMb(m.Sys))
  fmt.Println("NumGC = %v\n", m.NumGC)
  fmt.Println("--------------------------------")

  fmt.Println("\n")
}

func ReScrape(client *mongo.Client, ctx context.Context){
  cursor, _, err := database.GetFromCollection(client, ctx, "hosts", bson.D{})

  if err != nil{
    fmt.Println(err)
    return
  }

  if err := cursor.Err(); err != nil{
    fmt.Println(err)
    return
  }

  var urls []DatabaseEntry

  if err = cursor.All(ctx, &urls); err != nil {
    fmt.Println(err)
  }

  for x := range urls{
    go Execute("https://" + urls[x].Host, client, ctx)

    for y := range urls[x].URLS{
      go Execute(urls[x].URLS[y].Url, client, ctx)
    }
  }
}

func Execute(rawurl string, client *mongo.Client, ctx context.Context){
  u, err := url.Parse(rawurl)

  if err != nil{
    log.Fatal(err)
    return
  }

  urls, err := scrapeURL(rawurl)

  if err != nil{
    fmt.Println(err)
    return
  }

  var entries []DatabaseEntryURLS

  if urls == nil{
    return
  }

  for k := range urls.urls {
    p, _ := url.Parse(urls.urls[k].fullURL)

    // assumes it is running as test is client is nil
    if client != nil{
      go Execute(urls.urls[k].fullURL, client, ctx)
    }

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

  if client != nil{
    database.DeleteIfExists(client, ctx, "hosts", u.Host)

    for a := range entryNodes{
      _, err := database.AddToCollection(client, ctx, "hosts", entryNodes[a])

      if err != nil{
        log.Fatal(err)
      }
    }
  } else {
    fmt.Println(entryNodes)

    os.Exit(0)
  }

  // _, err = database.AddToCollectionAndDeleteIfExists("hosts", entryNodes, u.Host)

  // if err != nil {
  //   log.Fatal(err)
  // }

  // fmt.Println("Saving current data")
}


func getMeta(source string, rurl string) (*Meta, error){
  var name string
  var description string
  var icon string

  /*
    Meta Score - Determins what position your website is in the results

    +1 - Has a valid language meta (so the search engine can make sure it provides )

    +1 - The website is secure (uses https)

    +1 - The website has more than 20+ links linked in the page
  */


  score := 0
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

  language, e := doc.Find("html").Attr("lang")

  if e == false{
    language =  "en"
    
  } else {
    score += 1
  }

  l := doc.Find("a").Length()

  if l > 20{
    score += 1
  }

  secure := false

  p, _ := url.Parse(rurl)

  if p.Scheme == "https"{
    secure = true

    score += 1
  }

  result := &Meta{
    Name: name,
    Description: description,
    Icon: icon,
    Keywords: strings.Split(keywords, ","),
    Language: language,
    Secure: secure,
    Score: score,
  }

  return result, nil
}
 
func getRobotsTxt(rawurl string) (string ,error){
  u, err := url.Parse(rawurl)


  if err != nil{
    return "", err
  }

  rawurl = u.Scheme + "://" + u.Host + "/robots.txt"

  text, err := Request(rawurl)

  if err != nil{
    return "", err
  }
  
  return string(text), nil
}

func Request(rawurl string) (string, error){
  p, _ := url.Parse(rawurl)
  

  c, err := caching.Get("caching_" + p.Host + p.Path)

  // fmt.Println(err == nil)
  // fmt.Println(c == "")
  if c != ""{
    return c, nil
  }

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

  caching.Set("caching_" + p.Host + p.Path, string(body), time.Duration(2 * time.Hour))
  
  return string(body), nil
}

func scrapeURL(rawurl string) (*URLQuery, error){
  time.Sleep(time.Duration(5))
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

  meta, err := getMeta(string(body), request.URL.String())

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
          meta, err := getMeta(b, u.Scheme + "://" + u.Host + link)
          
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

      meta, err := getMeta(b, u.Scheme + "://" + u.Host + link)

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

      meta, err := getMeta(b, u.Scheme + "://" + u.Host + link)

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
