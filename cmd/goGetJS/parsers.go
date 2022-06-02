package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

func parseDoc(r io.Reader, baseURL string) ([]string, int, error) {
	scriptsSRC := []string{}
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return scriptsSRC, 0, fmt.Errorf("could not read HTML with goquery: %v", err)
	}

	anonCount := 0

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		// handling scripts with src
		if value, ok := s.Attr("src"); ok {
			if !strings.HasPrefix(value, "http") {
				if !strings.HasPrefix(value, "/") {
					value = fmt.Sprintf("/%s", value)
				}
				scriptsSRC = append(scriptsSRC, baseURL+value)
			} else {
				scriptsSRC = append(scriptsSRC, value)
			}
		} else {
			// handling scripts without src
			script := strings.TrimSpace(s.Text())

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

	return scriptsSRC, anonCount, fmt.Errorf("no src found at %v", baseURL)
}

func getJS(client *http.Client, url string, query interface{}, r *regexp.Regexp) error {
	log.Println("getting JavaScript from:", url)
	res, err := makeRequest(url, client)
	if err != nil {
		return fmt.Errorf("could not make request at %v: %v", url, err)
	}

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return fmt.Errorf("unable to parse %v: %v", url, err)
	}
	script := doc.Find("body").Text()

	switch q := query.(type) {
	case *regexp.Regexp:
		if q.FindAllString(script, -1) != nil {
			fmt.Printf("\n*** found %q in %s ***\n", q.FindAllString(script, -1), url)
		}

		if script != "" {
			err := writeScripts(script, url, r)
			if err != nil {
				return fmt.Errorf("unable to write script file: %v", err)
			}
			return nil
		}
	case string:
		if q != "" {
			if strings.Contains(script, q) {
				fmt.Printf("\n*** found %q in %s ***\n", q, url)
			}
		}
		if script != "" {
			err := writeScripts(script, url, r)
			if err != nil {
				return fmt.Errorf("unable to write script file: %v", err)
			}
			return nil
		}
	case []string:
		var g errgroup.Group
		for _, term := range q {
			t := term
			g.Go(func() error {
				if strings.Contains(script, t) {
					fmt.Printf("\n*** found %q in %s ***\n", t, url)
				}
				if script != "" {
					err := writeScripts(script, url, r)
					if err != nil {
						return fmt.Errorf("unable to write script file for %s: %v", url, err)
					}
					return nil
				}
				return nil
			})
		}
		err := g.Wait()
		if err != nil {
			fmt.Printf("error during search: %v", err)
		}
		fmt.Println("search completed")
		return nil

	default:
		return fmt.Errorf("malformed query, please try again: %v", err)
	}

	return fmt.Errorf("no scripts found at %v", url)
}
