package sitemaps

import (
  "github.com/yterajima/go-sitemap"
  "net/http"
  "io/ioutil"
  "time"
)


func GetURLS(s string) ([]string, error){
	sitemap.SetFetch(myFetch)

	smap, err := sitemap.Get(s, nil)

	if err != nil {
		return nil, err
	}

  var urls []string

	for _, URL := range smap.URL {
    urls = append(urls, URL.Loc)
	}

  return urls, nil
}

func myFetch(URL string, options interface{}) ([]byte, error) {
	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return []byte{}, err
	}

	// Set User-Agent
	req.Header.Set("User-Agent", "cosmic (+github.com/hackermondev/cosmic)")

	// Set timeout
	timeout := time.Duration(30 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	// Fetch data
	res, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return []byte{}, err
	}

	return body, err
}