package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/playwright-community/playwright-go"
	"golang.org/x/sync/errgroup"
)


func makeRequest(url string, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %v", err)
	}
	
	ua := randomUA()
	req.Header.Set("User-Agent", ua)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
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

func parseDoc(r io.Reader, baseURL string) ([]string, int, error) {
	scriptsSRC := []string{}	
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return scriptsSRC, 0, fmt.Errorf("could not read HTML with goquery: %v", err)
	}

	j := 0

	doc.Find("script").Each(func(i int, s *goquery.Selection) {	
		// scripts with src
		if value, ok := s.Attr("src"); ok {
			if !strings.HasPrefix(value, "http") {
				if !strings.HasPrefix(value, "/") {
					value = fmt.Sprintf("/%s", value)
				}
				scriptsSRC = append(scriptsSRC, baseURL + value)
			} else {
				scriptsSRC = append(scriptsSRC, value)
			}
		} else {
			// scripts without src
			script := strings.TrimSpace(s.Text())

			// write to file			
			scriptByte := []byte(script)
			j++
			scriptName := fmt.Sprintf("anon%s.js", strconv.Itoa(j))
			if err := os.WriteFile(scriptName, scriptByte, 0644); err != nil {
				log.Println("could not write anon script", err)
				j--
			}
		}
	})

	if len(scriptsSRC) != 0 {
		return scriptsSRC, j, nil
	}

	return scriptsSRC, j, fmt.Errorf("no src found on page")
}


func getJS(client *http.Client, url string) error {
	log.Println("getting script at:", url)
	res, err := makeRequest(url, client)
	if err != nil {
		return fmt.Errorf("could not make script request: %v", err)
	}
	err = parseScripts(res)
	if err != nil {
		return fmt.Errorf("no script available: %v", err)
	}
	
	return nil
}

func parseScripts(res *http.Response) error {
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return fmt.Errorf("unable to parse script page: %q", err)
	}
	script := doc.Find("body").Text()
	currentURL := *res.Request.URL
	url := currentURL.String()

	if script != "" {
		err := writeScripts(script, url)
		if err != nil {
			return fmt.Errorf("unable to write script file: %q", err)
		}
		return nil
	}
	
	return fmt.Errorf("no scripts at %v", url)
}

func writeScripts(script, url string) error {
	r := regexp.MustCompile(`[\w-]+(\.js)?$`) // need to expand? 
	fileName := r.FindString(url)
	scriptByte := []byte(script)
	if err := os.WriteFile(fileName, scriptByte, 0644); err != nil {
		return err
	}
	return nil
}

func getAbsURL(res *http.Response) string {
	base := *res.Request.URL
	abs := base.String()
	return strings.TrimSuffix(abs, "/")
}

func assertErrorToNilf(msg string, err error) {
	if err != nil {
		log.Fatalf(msg, err)
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

func noBrowser(url string, timeout int) {
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
			err := getJS(client, url)
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

func browser(url string, timeout int) {
	pw, err := playwright.Run()
	assertErrorToNilf("could not start playwright: %v", err)
	
	browser, err := pw.Chromium.Launch()
	assertErrorToNilf("could not launch useBrowswer: %v", err)

	ua := randomUA()
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(ua),
	})
	assertErrorToNilf("could not create context: %v", err)
	
	page, err := context.NewPage()
	assertErrorToNilf("could not create page: %v", err)
	
	_, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	assertErrorToNilf("could not go to %v:", err)
	
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

	g := new(errgroup.Group)
	for _, url := range scriptsSRC {
		url := url
		g.Go(func() error {
			err := getJS(client, url)
			return err
		})
	}

	counter = counter + len(scriptsSRC)

	if err := g.Wait(); err != nil {
		log.Println("error fetching script: ", err)
		counter --
	}

	fmt.Printf("\nsuccessfully wrote %d scripts\n", counter)
}

func main() {
	url := flag.String("url", "https://go.dev/", "url to get JavaScript from")
	timeout := flag.Int("timeout", 5, "timeout for request")
	useBrowswer := flag.Bool("useBrowser", false, "run playwright for JS-intensive sites (default is false")
	flag.Parse()

	start := time.Now()

	if !*useBrowswer {
		noBrowser(*url, *timeout)
	} else {
		browser(*url, *timeout)
	}
	
	fmt.Printf("\ntook: %f seconds\n", time.Since(start).Seconds())
}