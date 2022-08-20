package main

import (
	"fmt"
	"io"
	"net/http"
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
// found on the page, parseDoc writes the html to a file to aid in debugging.
func (app *application) parseDoc(r io.Reader, url string, query interface{}) ([]string, int, error) {
	var srcs []string
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return srcs, 0, fmt.Errorf("could not create goquery doc for %v: %w", url, err)
	}

	anonCount := 0

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		// handling scripts with src
		if src, ok := s.Attr("src"); ok {
			src = strings.TrimSpace(src)
			switch {
			case strings.HasPrefix(src, "//"):
				full := fmt.Sprintf("http:%s", src)
				srcs = append(srcs, full)
			case strings.HasPrefix(src, "/"):
				full := app.baseURL + src
				srcs = append(srcs, full)
			default:
				srcs = append(srcs, src)
			}
		} else {
			// handling scripts without src
			script := strings.TrimSpace(s.Text())

			// write scripts to file
			scriptByte := []byte(script)
			anonCount++
			scriptName := fmt.Sprintf("anon%s.js", strconv.Itoa(anonCount))
			app.searchScript(query, scriptName, script)
			if err := os.WriteFile("data/"+scriptName, scriptByte, 0644); err != nil {
				app.errorLog.Printf("could not write %q: %v", scriptName, err)
				anonCount--
			}
		}
	})

	if len(srcs) != 0 {
		return srcs, anonCount, nil
	}

	// if no src found, write the page to a file for debugging purposes
	html, err := doc.Html()
	if err != nil {
		return srcs, anonCount, fmt.Errorf("unable to get HTML for %v: %w", url, err)
	}
	err = app.writePage(html, url)
	if err != nil {
		return srcs, anonCount, fmt.Errorf("unable to write HTML for %v: %w", url, err)
	}

	return srcs, anonCount, fmt.Errorf("no src found at %v: %w", url, err)
}

// getJS takes in a url to a javascript file, extracts the contents, and writes them to an individual javascript file.
func (app *application) getJS(client *http.Client, url string, query interface{}, r *regexp.Regexp) error {
	app.infoLog.Println("extracting from:", url)
	resp, err := app.makeRequest(url, client)
	if err != nil {
		return fmt.Errorf("could not make request at %s: %w", url, err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read response body for %s: %w", url, err)
	}

	// retry (uses short timeout and allows redirects)
	if len(body) == 0 {
		go app.quickRetry(url, query, r)
	}

	script := string(body)

	app.searchScript(query, url, script)

	if script != "" {
		err := app.writeScript(script, url, r)
		if err != nil {
			return fmt.Errorf("unable to write script file: %w", err)
		}
		return nil
	}

	return nil
}

// searchScript takes a query, a url, and the script to be searched, and saves
// any found terms (and the url they were found on) to the SearchMap for
// later writing to a file
func (app *application) searchScript(query interface{}, url, script string) {
	switch q := query.(type) {
	case *regexp.Regexp:
		savedTerm := make(map[string]bool)
		if q.FindAllString(script, -1) != nil {
			for _, v := range q.FindAllString(script, -1) {
				if savedTerm[v] {
					continue
				}
				savedTerm[v] = true
				app.searches.Store(url, v)
			}
		}
	case string:
		if q != "" && strings.Contains(script, q) {
			app.searches.Store(url, q)
		}
	case []string:
		var wg sync.WaitGroup
		for _, term := range q {
			wg.Add(1)
			go func(t string) {
				if strings.Contains(script, t) {
					app.searches.Store(url, t)
				}
				wg.Done()
			}(term)
		}
		wg.Wait()
	default:
		app.errorLog.Println("malformed query, please try again")
	}
}
