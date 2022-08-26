package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// browser uses a headless browser (chromium via playwright) to scrape a site, waiting until there are no
// network connections for at least 500ms (unless a longer wait is requested with the extraWait) flag.
// browser returns an io.Reader and an error.
func (app *application) browser(url string, browserTimeout *float64, extraWait int, client *http.Client) (io.Reader, error) {
	fmt.Println("============================================================")
	app.infoLog.Println("initiating playwright browser...")

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("start playwright error: %w", err)
	}

	browser, err := pw.Chromium.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch browser error: %w", err)
	}

	uAgent := app.randomUA()
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(uAgent),
	})
	if err != nil {
		return nil, fmt.Errorf("create browser context error: %w", err)
	}

	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("create browser page error: %w", err)
	}

	_, err = page.Goto(url, playwright.PageGotoOptions{
		Timeout:   browserTimeout,
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return nil, fmt.Errorf("browser navigation error: %w", err)
	}

	if extraWait > 0 {
		time.Sleep(time.Duration(extraWait) * time.Millisecond)
		app.infoLog.Printf("slept for %d milliseconds\n", extraWait)
	}

	htmlDoc, err := page.Content()
	if err != nil {
		return nil, fmt.Errorf("playwright html extraction error: %w", err)
	}

	err = browser.Close()
	if err != nil {
		return nil, fmt.Errorf("close browser error: %w", err)
	}

	err = pw.Stop()
	if err != nil {
		return nil, fmt.Errorf("stop playwright error: %w", err)
	}

	app.infoLog.Println("browser finished")
	fmt.Println("============================================================")

	return strings.NewReader(htmlDoc), nil
}
