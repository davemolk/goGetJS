package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
	"golang.org/x/sync/errgroup"
)

func noBrowser(url, term string, timeout int) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err := makeRequest(url, client)
	assertErrorToNilf("could not make request: %v", err)
	defer res.Body.Close()

	baseURL := getAbsURL(res)

	// parse site
	scriptsSRC, counter, err := parseDoc(res.Body, baseURL)
	assertErrorToNilf("could not parse HTML: %v", err)

	// write to file
	err = writeFile(scriptsSRC, "scriptsSRC.txt")
	assertErrorToNilf("could not write src list to file: %v", err)

	// get JS
	g := new(errgroup.Group)
	for _, url := range scriptsSRC {
		url := url
		g.Go(func() error {
			err := getJS(client, url, term)
			return err
		})
	}

	counter = counter + len(scriptsSRC)

	if err := g.Wait(); err != nil {
		log.Println("error fetching script: ", err)
		counter--
	}

	fmt.Printf("\nsuccessfully wrote %d scripts\n", counter)
}

func browser(url, term string, timeout, extraWait int) {
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

	scriptsSRC, counter, err := parseDoc(strings.NewReader(htmlDoc), url)
	assertErrorToNilf("could not parse browser HTML: %v", err)

	err = writeFile(scriptsSRC, "scriptSRC.txt")
	assertErrorToNilf("could not write src list to file: %v", err)

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	group := new(errgroup.Group)
	for _, url := range scriptsSRC {
		url := url
		group.Go(func() error {
			err := getJS(client, url, term)
			return err
		})
	}

	counter = counter + len(scriptsSRC)

	if err := group.Wait(); err != nil {
		log.Println("error fetching script: ", err)
		counter--
	}

	fmt.Printf("\nsuccessfully wrote %d scripts\n", counter)
}
