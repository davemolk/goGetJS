package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

// makeClient takes in a flag-specified timeout and returns an *http.Client that has
// been configured according to the timeout, proxy, and redirect flags.
func (app *application) makeClient(timeout int, proxy string, redirect bool) *http.Client {
	switch {
	case proxy != "":
		parsed, err := url.Parse(proxy)
		if err != nil {
			app.errorLog.Fatalf("proxy error: %v", err)
		}

		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.Proxy = http.ProxyURL(parsed)

		return &http.Client{
			CheckRedirect: app.allowRedirects(redirect),
			Timeout:       time.Duration(timeout) * time.Millisecond,
			Transport:     tr,
		}
	default:
		return &http.Client{
			CheckRedirect: app.allowRedirects(redirect),
			Timeout:       time.Duration(timeout) * time.Millisecond,
		}
	}
}

// allowRedirects checks the redirects flag. If the flag is true, allowRedirects
// returns nil and redirects will be allowed. If the flag is false (the default value),
// allowRedirects returns a function for the CheckRedirect field of the http.Client that
// blocks redirects.
func (app *application) allowRedirects(redirect bool) func(req *http.Request, via []*http.Request) error {
	switch {
	case redirect:
		return nil
	default:
		return func(req *http.Request, via []*http.Request) error {
			app.errorLog.Printf("redirect to %s has been blocked\n", req.URL.String())
			return http.ErrUseLastResponse
		}
	}
}

// makeRequest takes in a url and a client, forms a new GET request, sets a random
// user agent, and then returns the response and any errors.
func (app *application) makeRequest(url string, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request error for %v: %w", url, err)
	}

	uAgent := app.randomUA()
	req.Header.Set("User-Agent", uAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("response error for %v: %w", url, err)
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("status code error for %v: %d", url, resp.StatusCode)
	}
	return resp, nil
}

// quickRetry uses a short timeout and allows redirects. It's called within getJS
// to retry any src that link to a page without text.
func (app *application) quickRetry(url string, query interface{}, r *regexp.Regexp) {
	resp, err := app.makeRequest(url, app.retryClient)
	if err != nil {
		app.errorLog.Printf("retry request error for %v: %v\n", url, err)
		resp.Body.Close()
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		app.errorLog.Printf("retry read error for %v: %v\n", url, err)
		return
	}
	script := string(b)

	if script != "" {
		app.infoLog.Printf("retry success: %v\n", url)
		app.searchScript(query, url, script)
		err := app.writeScript(script, url, r)
		if err != nil {
			app.errorLog.Printf("retry write error for %v: %v\n", url, err)
			return
		}
	}
}

// randomUA returns a user agent randomly drawn from six possibilities.
func (app *application) randomUA() string {
	userAgents := app.getUA()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	rando := r.Intn(len(userAgents))

	return userAgents[rando]
}

// getUA returns a string slice of six user agents.
func (app *application) getUA() []string {
	return []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/603.3.8 (KHTML, like Gecko)",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/601.7.7 (KHTML, like Gecko) Version/9.1.2 Safari/601.7.7",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:99.0) Gecko/20100101 Firefox/99.0",
	}
}
