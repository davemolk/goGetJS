package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)


type Site struct {
	ScriptSRC []string
	Scripts []string
}

func makeRequest(url string, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}
	
	ua := randomUA()
	req.Header.Set("User-Agent", ua)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func parseDoc(res *http.Response) ([]string, error) {
	var jsSRC []string

	site := Site{}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return jsSRC, fmt.Errorf("could not read HTML with goquery: %w", err)
	}

	baseURL := getAbs(res)

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if value, ok := s.Attr("src"); ok {
			if !strings.HasPrefix(value, "http") {
				jsSRC = append(jsSRC, baseURL + value)
			} else {
				jsSRC = append(jsSRC, value)
			}
		}
	})

	// TODO grab script content where there is no src
	site.ScriptSRC = jsSRC
	if len(jsSRC) != 0 {
		return jsSRC, nil
	}

	return jsSRC, fmt.Errorf("no src found on page")
}

func getJS(client *http.Client, scriptSRC []string) ([]string, error) {
	var jsScripts []string
	for _, s := range scriptSRC {
		log.Println("script name is:", s)
		res, err := makeRequest(s, client)
		if err != nil {
			return jsScripts, fmt.Errorf("could not make script request: %s", err)
		}
		script, err := parseScripts(res)
		if err != nil {
			log.Printf("no script available: %s\n", err)
		}
		jsScripts = append(jsScripts, script)
	}
	return jsScripts, nil
}

func parseScripts(res *http.Response) (string, error) {
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("unable to parse script page: %s", err)
	}
	script := doc.Find("body").Text()
	currentURL := *res.Request.URL
	url := currentURL.String()

	if script != "" {
		err := writeScripts(script, url) // flag for this
		if err != nil {
			return script, fmt.Errorf("unable to write script file: %s", err)
		}
		return script, nil
	}
	
	return "", fmt.Errorf("no scripts at %s", url)
}

func writeScripts(script, url string) error {
	r := regexp.MustCompile(`[\w-]+(\.js)?$`)
	fileName := r.FindString(url)
	scriptByte := []byte(script)
	if err := os.WriteFile(fileName, scriptByte, 0644); err != nil {
		return err
	}
	return nil
}

func getAbs(res *http.Response) string {
	base := *res.Request.URL
	abs := base.String()
	return strings.TrimSuffix(abs, "/")
}

func assertErrorToNilf(msg string, err error) {
	if err != nil {
		log.Fatalf(msg, err)
	}
}

func randomUA() string {
	userAgents := getUA()
	r := rand.New(rand.NewSource(time.Now().Unix()))
	rando := r.Intn(len(userAgents))

	return userAgents[rando]
}

func getUA() []string {
	return []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/603.3.8 (KHTML, like Gecko)",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/601.7.7 (KHTML, like Gecko) Version/9.1.2 Safari/601.7.7",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:99.0) Gecko/20100101 Firefox/99.0",
	}
}

func writeFile(scripts []string, fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, v := range scripts {
		fmt.Fprintln(f, v)
	}
	return nil
}

func main() {
	url := flag.String("url", "https://go.dev/", "url to get JavaScript from")
	timeout := flag.Int("timeout", 5, "timeout for request")
	pw := flag.Bool("pw", false, "run playwright for JS-intensive sites (default is false")
	flag.Parse()

	start := time.Now()

	_ = pw 

	client := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err := makeRequest(*url, client)
	assertErrorToNilf("could not make request: %s", err)

	site := Site{}
	// parse
	scriptSRC, err := parseDoc(res)
	assertErrorToNilf("could not parse HTML: %s", err)

	site.ScriptSRC = scriptSRC

	// write to file (add a flag for this)
	err = writeFile(scriptSRC, "jsLinks.txt")
	assertErrorToNilf("could not write to file: %s", err)
	
	// get JS // goroutines
	jsScripts, err := getJS(client, scriptSRC)
	assertErrorToNilf("could not make script request: %s", err)
	site.Scripts = jsScripts

	// write to file // goroutines
	// err = writeFile(jsScripts, "JS.txt")
	// assertErrorToNilf("could not write to file: %s", err)

	fmt.Printf("\ntook: %f seconds\n", time.Since(start).Seconds())
}