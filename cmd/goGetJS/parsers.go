package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// parseDoc searches a page for script tags, returning a string slice of all found src, the
// number found, and any errors. When a script tag does not have an src attribute, parseDoc
// writes the contents between the script tags as an anonymous javascript file. If no src are
// found on the page, parseDoc returns the html as a string to aid in debugging.
func parseDoc(r io.Reader, myUrl string, query interface{}) ([]string, int, error) {
	srcs := []string{}
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return srcs, 0, fmt.Errorf("could not create goquery doc for %v: %v", myUrl, err)
	}

	anonCount := 0

	u, err := url.Parse(myUrl)
	if err != nil {
		return srcs, 0, fmt.Errorf("could not parse %v URL: %v", myUrl, err)
	}

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		// handling scripts with src
		if src, ok := s.Attr("src"); ok {
			src = strings.TrimSpace(src)
			if !strings.HasPrefix(src, "http") {
				rel, err := u.Parse(src)
				if err != nil {
					log.Printf("unable to parse %v (found on %v):\n%v", src, myUrl, err)
				}
				srcs = append(srcs, rel.String())
			} else {
				srcs = append(srcs, src)
			}
		} else {
			// handling scripts without src
			script := strings.TrimSpace(s.Text())

			searchScript(query, myUrl, script)

			// write scripts to file
			scriptByte := []byte(script)
			anonCount++
			scriptName := fmt.Sprintf("anon%s.js", strconv.Itoa(anonCount))
			if err := os.WriteFile("data/"+scriptName, scriptByte, 0644); err != nil {
				log.Printf("could not write %q: %v", scriptName, err)
				anonCount--
			}
		}
	})

	if len(srcs) != 0 {
		return srcs, anonCount, nil
	}

	// no src found
	html, _ := doc.Html()
	return srcs, anonCount, fmt.Errorf("no src found at %v\nif your url isn't the root domain, consider adding/removing a trailing slash\n%v", myUrl, html)
}

// getJS retrieves and then writes the contents a url (in this case, each src) to an individual javascript file.
func getJS(client *http.Client, url string, query interface{}, r *regexp.Regexp) error {
	log.Println("extracting JavaScript from:", url)
	resp, err := makeRequest(url, client)
	if err != nil {
		return fmt.Errorf("could not make request at %s: %v", url, err)
	}

	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("could not create goquery doc for %s: %v", url, err)
	}

	script := doc.Find("body").Text()
	searchScript(query, url, script)

	if script != "" {
		err := writeScript(script, url, r)
		if err != nil {
			return fmt.Errorf("unable to write script file: %v", err)
		}
		return nil
	}

	return fmt.Errorf("no scripts found at %v", url)
}

// searchScript takes a query (as an empty interface), a url, and the script to be searched,
// printing any found instances to the console.
func searchScript(query interface{}, url, script string) {
	switch q := query.(type) {
	case *regexp.Regexp:
		if q.FindAllString(script, -1) != nil {
			fmt.Printf("\n*** found %q in %s ***\n", q.FindAllString(script, -1), url)
		}
	case string:
		if q != "" && strings.Contains(script, q) {
			fmt.Printf("\n*** found %q in %s ***\n", q, url)
		}
	case []string:
		var wg sync.WaitGroup
		for _, term := range q {
			wg.Add(1)
			go func(t string) {
				if strings.Contains(script, t) {
					fmt.Printf("\n*** found %q in %s ***\n", t, url)
				}
				wg.Done()
			}(term)
		}
		wg.Wait()
	default:
		fmt.Println("malformed query, please try again")
	}
}
