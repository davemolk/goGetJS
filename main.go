package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)


func makeRequest(url string, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}
	// req.Header.Set("User-Agent", uAgent)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func parseDoc(res *http.Response) ([]string, error) {
	var jsSRC []string

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return jsSRC, fmt.Errorf("could not read HTML with goquery: %w", err)
	}

	// eventually add complete urls...
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if value, ok := s.Attr("src"); ok {
			jsSRC = append(jsSRC, value)
		}
	})
	if jsSRC != nil {
		return jsSRC, nil
	}

	return jsSRC, fmt.Errorf("no src found on page")

}

func assertErrorToNilf(msg string, err error) {
	if err != nil {
		log.Fatalf(msg, err)
	}
}

func main() {
	url := flag.String("url", "https://go.dev/", "url to get JavaScript from")
	timeout := flag.Int("timeout", 5, "timeout for request")
	pw := flag.Bool("pw", false, "run playwright for JS-intensive sites (default is false")
	flag.Parse()

	_ = pw 

	client := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err := makeRequest(*url, client)
	assertErrorToNilf("could not launch browser: %s", err)

	// parse
	scripts, err := parseDoc(res)
	assertErrorToNilf("could not parse HTML: %s", err)
	fmt.Println("scripts:", scripts)

	

}

