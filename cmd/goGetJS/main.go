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
	var reg string

	url := flag.String("url", "https://go.dev", "url to get JavaScript from")
	timeout := flag.Int("timeout", 5, "timeout for request")
	useBrowswer := flag.Bool("browser", false, "run playwright for JS-intensive sites (default is false")
	extraWait := flag.Int("extraWait", 0, "wait (in seconds) for longer network events, only applies when browser=true. default is 0 seconds")
	flag.StringVar(&term, "term", "", "search term")
	flag.StringVar(&reg, "regex", "", "regex expression")

	flag.Parse()

	err := os.Mkdir("data", 0755)
	assertErrorToNilf("could not create folder to store scripts: %v", err)

	start := time.Now()

	client := makeClient(*timeout)

	var reader io.Reader

	if *useBrowswer {
		reader, err = browser(*url, term, *extraWait, client)
		if err != nil {
			log.Println("error for use browser")
		}		
	} else {
		res, err := makeRequest(*url, client)
		assertErrorToNilf("could not make request: %v", err)
		reader = res.Body
		defer res.Body.Close()
	}

	scriptsSRC, counter, err := parseDoc(reader, *url)
	assertErrorToNilf("could not parse browser HTML: %v", err)

	err = writeFile(scriptsSRC, "scriptSRC.txt")
	assertErrorToNilf("could not write src list to file: %v", err)
	
	var query interface{}
	if len(reg) > 0 {
		re := regexp.MustCompile(reg)
		query = re // need to compile here...
	} else {
		query = term
	}

	group := new(errgroup.Group)
	for _, url := range scriptsSRC {
		url := url
		group.Go(func() error {
			err := getJS(client, url, query)
			return err
		})
	}

	counter = counter + len(scriptsSRC)

	if err := group.Wait(); err != nil {
		log.Println("error fetching script: ", err)
		counter--
	}

	fmt.Printf("\nsuccessfully wrote %d scripts\n", counter)

	fmt.Printf("\ntook: %f seconds\n", time.Since(start).Seconds())
}