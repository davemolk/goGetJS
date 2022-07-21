package main

import (
	"bufio"
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
	inputFile      string
	regex          string
	term           string
	timeout        int
	url            string
	useBrowser     bool
}

type application struct {
	baseURL string
	client   *http.Client
	config   config
	errorLog *log.Logger
	infoLog  *log.Logger
	searchLog *log.Logger
}

func main() {
	var cfg config

	flag.StringVar(&cfg.url, "u", "", "url for getting JavaScript")
	flag.IntVar(&cfg.timeout, "t", 5, "timeout for request")
	flag.Float64Var(&cfg.browserTimeout, "bt", 10000, "browser timeout")
	flag.BoolVar(&cfg.useBrowser, "b", false, "use playwright to handle JS-intensive sites (default is false")
	flag.IntVar(&cfg.extraWait, "ew", 0, "additional wait (in seconds) when using a browser. default is 0 seconds")
	flag.StringVar(&cfg.term, "w", "", "search JavaScript for a particular word")
	flag.StringVar(&cfg.regex, "r", "", "search JavaScript with a regex expression")
	flag.StringVar(&cfg.inputFile, "f", "", "file containing a list of search terms")

	flag.Parse()

	start := time.Now()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ltime|log.Lshortfile)
	searchLog := log.New(os.Stdout, "FOUND\t", log.Ltime)

	app := &application{
		config:   cfg,
		errorLog: errorLog,
		infoLog:  infoLog,
		searchLog: searchLog,
	}

	stat, err := os.Stdin.Stat()
	app.assertErrorToNilf("Stdin path error: %v", err)

	if cfg.url == "" {
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			s := bufio.NewScanner(os.Stdin)
			for s.Scan() {
				cfg.url = s.Text()
			}
		}
	}

	if cfg.url == "" {
		app.errorLog.Fatal("must provide a URL")
	}

	// maybe use a list to filter out .dev, .org, etc from this?
	baseURL, err := app.getBaseURL(cfg.url)
	app.assertErrorToNilf("unable to parse base URL", err)

	app.baseURL = baseURL

	err = os.Mkdir("data", 0755)
	app.assertErrorToNilf("could not create folder to store javascript: %v", err)

	app.client = app.makeClient(cfg.timeout)

	var reader io.Reader

	// get reader
	if cfg.useBrowser {
		reader, err = app.browser(cfg.url, &cfg.browserTimeout, cfg.extraWait, app.client)
		app.assertErrorToNilf("could not make request with browser: %v", err)
	} else {
		resp, err := app.makeRequest(cfg.url, app.client)
		app.assertErrorToNilf("could not make request: %v", err)
		defer resp.Body.Close()
		reader = resp.Body
	}

	// configure query
	var query interface{}

	switch {
	case len(cfg.regex) > 0:
		re := regexp.MustCompile(cfg.regex)
		query = re
	case len(cfg.inputFile) > 0:
		f, err := os.Open(cfg.inputFile)
		app.assertErrorToNilf("could not open input file: %v", err)
		defer f.Close()

		lines, err := app.readLines(f)
		app.assertErrorToNilf("could not read input file: %v", err)
		query = lines
	default:
		query = cfg.term
	}

	// parse for src, writing javascript files without src
	srcs, anonCount, err := app.parseDoc(reader, cfg.url, query)
	app.assertErrorToNilf("could not parse HTML: %v", err)

	// write src text file
	err = app.writeFile(srcs, "scriptSRC.txt")
	app.assertErrorToNilf("could not write scriptSRC.txt: %v", err)

	// handling situations when src doesn't end with .js
	fName := regexp.MustCompile(`[\w-&]+(\.js)?$`)

	// extract, search, and write javascript files with src
	var g errgroup.Group
	for _, src := range srcs {
		src := src
		g.Go(func() error {
			err := app.getJS(app.client, src, query, fName)
			if err != nil {
				return fmt.Errorf("error while processing %v: %v", src, err)
			}
			return nil
		})
	}

	counter := anonCount + len(srcs)

	if err := g.Wait(); err != nil {
		app.errorLog.Printf("error with extract/search/write: %v", err)
		counter--
	}

	fmt.Println()
	fmt.Println("============================================================")
	app.infoLog.Printf("successfully processed %d scripts\n", counter)
	app.infoLog.Printf("took: %f seconds\n", time.Since(start).Seconds())
	fmt.Println("============================================================")
}
