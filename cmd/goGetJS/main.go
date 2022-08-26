package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"golang.org/x/sync/errgroup"
)

type config struct {
	browserTimeout float64
	extraWait      int
	proxy          string
	redirect       bool
	regex          string
	retryTimeout   int
	term           string
	terms          string
	timeout        int
	url            string
	useBrowser     bool
}

type application struct {
	baseURL     string
	client      *http.Client
	config      config
	errorLog    *log.Logger
	infoLog     *log.Logger
	query       interface{}
	retryClient *http.Client
	searches    *SearchMap
}

func main() {
	var cfg config

	flag.Float64Var(&cfg.browserTimeout, "bt", 10000, "browser timeout (in ms). default 10000")
	flag.IntVar(&cfg.extraWait, "ew", 0, "additional wait (in ms) when using a browser. default 0")
	flag.StringVar(&cfg.proxy, "proxy", "", "enter a proxy to use")
	flag.BoolVar(&cfg.redirect, "redirect", false, "allow redirects. default is false")
	flag.StringVar(&cfg.regex, "regex", "", "search JavaScript with a regex expression")
	flag.IntVar(&cfg.retryTimeout, "rt", 1000, "timeout for retries. default is 1000ms")
	flag.StringVar(&cfg.term, "term", "", "search JavaScript for a particular term")
	flag.StringVar(&cfg.terms, "terms", "", "upload a file containing a list of search terms")
	flag.IntVar(&cfg.timeout, "t", 5000, "timeout (in ms) for request. default 5000")
	flag.StringVar(&cfg.url, "u", "", "url for getting JavaScript")
	flag.BoolVar(&cfg.useBrowser, "b", false, "use playwright to handle JS-intensive sites. default false")

	flag.Parse()

	start := time.Now()

	errorLog := log.New(os.Stderr, "ERROR\t", log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime)
	searches := NewSearchMap()

	app := &application{
		config:   cfg,
		errorLog: errorLog,
		infoLog:  infoLog,
		searches: searches,
	}

	app.getQuery()

	if app.config.url == "" {
		err := app.getInput()
		app.assertErrorToNilf(err)
	}

	baseURL, err := app.getBaseURL(cfg.url)
	app.assertErrorToNilf(err)
	app.baseURL = baseURL

	err = os.Mkdir("data", 0755)
	app.assertErrorToNilf(err)

	if cfg.term != "" || cfg.terms != "" || cfg.regex != "" {
		err := os.Mkdir("searchResults", 0755)
		app.assertErrorToNilf(err)
	}

	app.client = app.makeClient(cfg.timeout, cfg.proxy, cfg.redirect)
	app.retryClient = app.makeClient(cfg.retryTimeout, cfg.proxy, true)

	var reader io.Reader

	// get reader
	switch {
	case cfg.useBrowser:
		reader, err = app.browser(cfg.url, &cfg.browserTimeout, cfg.extraWait, app.client)
		app.assertErrorToNilf(err)
	default:
		resp, err := app.makeRequest(cfg.url, app.client)
		app.assertErrorToNilf(err)
		defer resp.Body.Close()
		reader = resp.Body
	}

	// parse for src, writing javascript files without src
	srcs, anonCount, err := app.parseDoc(reader, cfg.url, app.query)
	app.assertErrorToNilf(err)

	// write src text file
	err = app.writeFile(srcs, "scriptSRC.txt")
	app.assertErrorToNilf(err)

	// handling situations when src doesn't end with .js
	fName := regexp.MustCompile(`[\w-&]+(\.js)?$`)

	// extract, search, and write javascript files with src
	var g errgroup.Group
	for _, src := range srcs {
		src := src
		g.Go(func() error {
			err := app.getJS(app.client, src, app.query, fName)
			if err != nil {
				return err
			}
			return nil
		})
	}

	counter := anonCount + len(srcs)

	if err := g.Wait(); err != nil {
		app.errorLog.Printf("extract/search/write error: %v", err)
		counter--
	}

	// save search results (if applicable)
	if cfg.term != "" || cfg.terms != "" || cfg.regex != "" {
		err = app.writeSearchResults(app.searches.Searches)
		app.assertErrorToNilf(err)
	}

	fmt.Println()
	fmt.Println("============================================================")
	app.infoLog.Printf("successfully processed %d scripts\n", counter)
	app.infoLog.Printf("took %f seconds\n", time.Since(start).Seconds())
	fmt.Println("============================================================")
}
