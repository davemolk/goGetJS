package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	var term string
	var regex string
	var inputFile string

	url := flag.String("url", "https://go.dev", "url for getting JavaScript")
	timeout := flag.Int("timeout", 5, "timeout for request")
	useBrowswer := flag.Bool("browser", false, "use playwright to handle JS-intensive sites (default is false")
	extraWait := flag.Int("extraWait", 0, "additional wait (in seconds) when using a browser. default is 0 seconds")
	flag.StringVar(&term, "term", "", "search JavaScript for a particular term")
	flag.StringVar(&regex, "regex", "", "search JavaScript with a regex expression")
	flag.StringVar(&inputFile, "file", "", "file containing a list of search terms")

	flag.Parse()

	err := os.Mkdir("data", 0755)
	assertErrorToNilf("could not create folder to store javascript: %v", err)

	start := time.Now()

	client := makeClient(*timeout)

	var reader io.Reader

	// get reader
	if *useBrowswer {
		reader, err = browser(*url, *extraWait, client)
		assertErrorToNilf("could not make request with browser: %v", err)
	} else {
		resp, err := makeRequest(*url, client)
		assertErrorToNilf("could not make request: %v", err)
		reader = resp.Body
		defer resp.Body.Close()
	}

	// configure query
	var query interface{}

	if len(regex) > 0 {
		re := regexp.MustCompile(regex)
		query = re
	} else if len(inputFile) > 0 {
		f, err := os.Open(inputFile)
		assertErrorToNilf("could not open input file: %v", err)
		defer f.Close()

		lines, err := readLines(f)
		assertErrorToNilf("could not read input file: %v", err)
		query = lines
	} else {
		query = term
	}

	// parse for src, writing javascript files without src
	srcs, anonCount, err := parseDoc(reader, *url, query)
	assertErrorToNilf("could not parse HTML: %v", err)

	// write src text file
	err = writeFile(srcs, "scriptSRC.txt")
	assertErrorToNilf("could not write scriptSRC.txt: %v", err)

	// handling situations when src doesn't end with .js
	fName := regexp.MustCompile(`[\w-]+(\.js)?$`)

	// extract, search, and write javascript files with src
	var g errgroup.Group
	for _, src := range srcs {
		src := src
		g.Go(func() error {
			err := getJS(client, src, query, fName)
			if err != nil {
				return fmt.Errorf("error while processing %v: %v", src, err)
			}
			return nil
		})
	}

	counter := anonCount + len(srcs)

	if err := g.Wait(); err != nil {
		log.Printf("error with extract/search/write: %v", err)
		counter--
	}

	fmt.Printf("\nsuccessfully wrote %d scripts\n", counter)

	fmt.Printf("\ntook: %f seconds\n", time.Since(start).Seconds())
}
