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

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

// parseDoc searches a page for script tags, returning a string slice of all found src, the
// number found, and any errors. When a script tag does not have an src attribute, parseDoc
// writes the contents between the script tags as an anonymous javascript file. If no src are
// found on the page, parseDoc returns the html as a string to aid in debugging.
func parseDoc(r io.Reader, myUrl string, query interface{}) ([]string, int, error) {
	scriptsSRC := []string{}
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return scriptsSRC, 0, fmt.Errorf("could not read HTML with goquery: %v", err)
	}

	anonCount := 0

	u, err := url.Parse(myUrl)
	if err != nil {
		return scriptsSRC, 0, fmt.Errorf("unable to parse URL: %v", err)
	}

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		// handling scripts with src
		if value, ok := s.Attr("src"); ok {
			if !strings.HasPrefix(value, "http") {
				rel, err := u.Parse(value)
				if err != nil {
					log.Printf("unable to parse %v: \n%v\n", value, err)
				}
				scriptsSRC = append(scriptsSRC, rel.String())
			} else {
				scriptsSRC = append(scriptsSRC, value)
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
				log.Printf("could not write anon script %q: %v", scriptName, err)
				anonCount--
			}
		}
	})

	if len(scriptsSRC) != 0 {
		return scriptsSRC, anonCount, nil
	}

	// return html if no src is found
	html, _ := doc.Html()
	return scriptsSRC, anonCount, fmt.Errorf("no src found at %v\nif your url isn't the root domain, consider adding/removing a trailing slash\n%v", myUrl, html)
}

// getJS retrieves and then writes the contents a url (in this case, each src) to an individual javascript file.
func getJS(client *http.Client, url string, query interface{}, r *regexp.Regexp) error {
	log.Println("getting JavaScript from:", url)
	res, err := makeRequest(url, client)
	if err != nil {
		return fmt.Errorf("could not make request at %s: %v", url, err)
	}

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return fmt.Errorf("unable to parse %s: %v", url, err)
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
		if q != "" {
			if strings.Contains(script, q) {
				fmt.Printf("\n*** found %q in %s ***\n", q, url)
			}
		}
	case []string:
		var g errgroup.Group
		for _, term := range q {
			t := term
			g.Go(func() error {
				if strings.Contains(script, t) {
					fmt.Printf("\n*** found %q in %s ***\n", t, url)
				}
				return nil
			})
		}
		err := g.Wait()
		if err != nil {
			fmt.Printf("error during search: %v", err)
		}
	default:
		fmt.Println("malformed query, please try again")
	}
}
