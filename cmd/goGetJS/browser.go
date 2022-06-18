package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// browser uses a headless browser (playwright) to scrape a site, waiting until there are no
// network connections for at least 500ms (unless a longer wait is requested via the extraWait) flag.
// browser returns an io.Reader and an error.
func browser(url string, browserTimeout *float64, extraWait int, client *http.Client) (io.Reader, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch()
	if err != nil {
		return nil, fmt.Errorf("could not launch browswer: %v", err)
	}

	uAgent := randomUA()
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(uAgent),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create browser context: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create browser page: %v", err)
	}

	_, err = page.Goto(url, playwright.PageGotoOptions{
		Timeout: browserTimeout,
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return nil, fmt.Errorf("could not go to %v: %v", url, err)
	}

	if extraWait > 0 {
		time.Sleep(time.Duration(extraWait) * time.Second)
		log.Println("slept for", extraWait)
	}

	htmlDoc, err := page.Content()
	if err != nil {
		return nil, fmt.Errorf("could not get html from playwright: %v", err)
	}

	err = browser.Close()
	if err != nil {
		return nil, fmt.Errorf("could not close browser: %v", err)
	}

	err = pw.Stop()
	if err != nil {
		return nil, fmt.Errorf("could not stop playwright: %v", err)
	}

	return strings.NewReader(htmlDoc), nil
}
