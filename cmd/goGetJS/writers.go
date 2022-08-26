package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// writeScript writes a passed in string of javascript to an individual file.
func (app *application) writeScript(script, url string, fileNamer *regexp.Regexp) error {
	fName := fileNamer.FindString(url)
	url = "// " + url + "\n"
	urlByte := []byte(url)
	scriptByte := []byte(script)
	data := append(urlByte, scriptByte...)
	if err := os.WriteFile("data/"+fName, data, 0644); err != nil {
		return fmt.Errorf("cannot write %q: %w", fName, err)
	}
	return nil
}

// writeFile takes a string slice of src and writes the contents to a text file.
func (app *application) writeFile(scripts []string, fName string) error {
	f, err := os.Create(fName)
	if err != nil {
		return fmt.Errorf("file creation error: %w", err)
	}
	defer f.Close()
	for _, v := range scripts {
		if _, err := fmt.Fprintln(f, v); err != nil {
			return fmt.Errorf("file write error: %w", err)
		}
	}
	return nil
}

// writePage takes the html of a page (as a string) and the url and writes
// the contents to a text file.
func (app *application) writePage(s, myURL string) error {
	var n string // for naming purposes

	myURL = strings.TrimSuffix(myURL, "/")
	u, err := url.Parse(myURL)
	if err != nil {
		return fmt.Errorf("parse error for %v: %w", myURL, err)
	}

	switch {
	case u.Path != "":
		re := regexp.MustCompile(`[-\w\?\=]+/?$`)
		n = re.FindString(u.Path)
		n = strings.TrimSuffix(n, "/")
	default:
		n = u.Host
		n = strings.ReplaceAll(n, ".", "")
	}

	err = os.Mkdir("debug", 0755)
	if err != nil {
		return fmt.Errorf("debug folder creation error: %w", err)
	}
	f, err := os.Create("debug/" + n + ".html")
	if err != nil {
		return fmt.Errorf("debug file creation error for %v: %w", myURL, err)
	}
	defer f.Close()
	_, err = f.WriteString(s)
	if err != nil {
		return fmt.Errorf("write file error for %v: %w", myURL, err)
	}
	err = f.Sync()
	if err != nil {
		return fmt.Errorf("sync error for %v: %w", myURL, err)
	}
	return nil
}

// writeSearchResults takes in the search results (saved as an instance of SearchMap),
// creates a searchResults directory, and writes the results as a json file.
func (app *application) writeSearchResults(data map[string][]string) error {
	var n string

	switch app.query.(type) {
	case *regexp.Regexp:
		n = "regexResults.json"
	case string:
		n = "termResults.json"
	case []string:
		n = "termsResults.json"
	}

	sr, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("search results marshal error: %w", err)
	}
	f, err := os.Create(fmt.Sprintf("searchResults/%v", n))
	if err != nil {
		return fmt.Errorf("search results file error: %w", err)
	}
	defer f.Close()
	_, err = f.Write(sr)
	if err != nil {
		return fmt.Errorf("search results write error: %w", err)
	}
	err = f.Sync()
	if err != nil {
		return fmt.Errorf("search results sync error: %w", err)
	}
	return nil
}
