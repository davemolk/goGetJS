package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	url := flag.String("url", "https://go.dev/", "url to get JavaScript from")
	timeout := flag.Int("timeout", 5, "timeout for request")
	useBrowswer := flag.Bool("browser", false, "run playwright for JS-intensive sites (default is false")
	extraWait := flag.Int("extraWait", 0, "wait (in seconds) for longer network events, only applies when browser=true. default is 0 seconds")
	flag.Parse()

	err := os.Mkdir("data", 0755)
	assertErrorToNilf("could not create folder to store scripts: %v", err)

	start := time.Now()

	if !*useBrowswer {
		noBrowser(*url, *timeout)
	} else {
		browser(*url, *timeout, *extraWait)
	}

	fmt.Printf("\ntook: %f seconds\n", time.Since(start).Seconds())
}
