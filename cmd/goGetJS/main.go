package main

import (
	"bufio"
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
	var url string
	var term string
	var regex string
	var inputFile string
	flag.StringVar(&url, "u", "", "url for getting JavaScript")
	timeout := flag.Int("t", 5, "timeout for request")
	browserTimeout := flag.Float64("bt", 10000, "browser timeout")
	useBrowswer := flag.Bool("b", false, "use playwright to handle JS-intensive sites (default is false")
	extraWait := flag.Int("ew", 0, "additional wait (in seconds) when using a browser. default is 0 seconds")
	flag.StringVar(&term, "w", "", "search JavaScript for a particular word")
	flag.StringVar(&regex, "r", "", "search JavaScript with a regex expression")
	flag.StringVar(&inputFile, "f", "", "file containing a list of search terms")

	flag.Parse()

	start := time.Now()

	stat, err := os.Stdin.Stat()
	assertErrorToNilf("Stdin path error: %v", err)

	if url == "" {
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			s := bufio.NewScanner(os.Stdin)
			for s.Scan() {
				url = s.Text()
			}
		}
	}

	if url == "" {
		log.Fatal("must provide a URL")
	}

	err = os.Mkdir("data", 0755)
	assertErrorToNilf("could not create folder to store javascript: %v", err)
	client := makeClient(*timeout)

	var reader io.Reader

	// get reader
	if *useBrowswer {
		reader, err = browser(url, browserTimeout, *extraWait, client)
		assertErrorToNilf("could not make request with browser: %v", err)
	} else {
		resp, err := makeRequest(url, client)
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
	srcs, anonCount, err := parseDoc(reader, url, query)
	assertErrorToNilf("could not parse HTML: %v", err)

	// write src text file
	err = writeFile(srcs, "scriptSRC.txt")
	assertErrorToNilf("could not write scriptSRC.txt: %v", err)

	// handling situations when src doesn't end with .js
	fName := regexp.MustCompile(`[\w-&]+(\.js)?$`)

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

	fmt.Printf("\nsuccessfully processed %d scripts\n", counter)

	fmt.Printf("\ntook: %f seconds\n", time.Since(start).Seconds())
}
