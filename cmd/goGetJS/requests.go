package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"time"
)

// makeClient takes in a flag-specified timeout and returns an *http.Client.
func makeClient(timeout int) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// makeRequest takes in a url and a client, forms a new GET request, sets a random
// user agent, and then returns the response and any errors.
func makeRequest(url string, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request for %v: %v", url, err)
	}

	uAgent := randomUA()
	req.Header.Set("User-Agent", uAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to get response for %v: %v", url, err)
	}
	return resp, nil
}

// quickRetry uses a short timeout and allows redirects. It's called within getJS
// to retry any src showing no text on the page.
func quickRetry(url string, query interface{}, r *regexp.Regexp) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Printf("unable to create request for retry of %s: %v", url, err)
		return
	}

	uAgent := randomUA()
	req.Header.Set("User-Agent", uAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("unable to get response for retry of %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	script := string(b)

	if script != "" {
		log.Printf("retry for %s was successful\n", url)
		searchScript(query, url, script)
		writeScript(script, url, r)
		if err != nil {
			log.Printf("retry for %s was unsuccessful: unable to write script file: %v", url, err)
			return
		}
	}
}

// randomUA returns a user agent randomly drawn from six possibilities.
func randomUA() string {
	userAgents := getUA()
	r := rand.New(rand.NewSource(time.Now().Unix()))
	rando := r.Intn(len(userAgents))

	return userAgents[rando]
}

// getUA returns a string slice of six user agents.
func getUA() []string {
	return []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/603.3.8 (KHTML, like Gecko)",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/601.7.7 (KHTML, like Gecko) Version/9.1.2 Safari/601.7.7",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:99.0) Gecko/20100101 Firefox/99.0",
	}
}
