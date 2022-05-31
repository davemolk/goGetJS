package main

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

func browser(url, term string, extraWait int, client *http.Client) (io.Reader, error) {
	url = strings.TrimSuffix(url, "/")

	pw, err := playwright.Run()
	assertErrorToNilf("could not start playwright: %v", err)

	browser, err := pw.Chromium.Launch()
	assertErrorToNilf("could not launch useBrowswer: %v", err)

	uAgent := randomUA()
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(uAgent),
	})
	assertErrorToNilf("could not create context: %v", err)

	page, err := context.NewPage()
	assertErrorToNilf("could not create page: %v", err)

	_, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	assertErrorToNilf("could not go to %v:", err)

	time.Sleep(time.Duration(extraWait) * time.Second)

	htmlDoc, err := page.Content()
	assertErrorToNilf("could not get html from playwright: %v", err)

	return strings.NewReader(htmlDoc), nil
}
