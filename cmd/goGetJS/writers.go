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
		return fmt.Errorf("cannot write script %q: %w", fName, err)
	}
	return nil
}

// writeFile takes a string slice of src and writes the contents to a text file.
func (app *application) writeFile(scripts []string, fName string) error {
	f, err := os.Create(fName)
	if err != nil {
		return fmt.Errorf("could not create %q: %w", fName, err)
	}
	defer f.Close()
	for _, v := range scripts {
		if _, err := fmt.Fprintln(f, v); err != nil {
			return fmt.Errorf("unable to write %q in %q: %w", v, fName, err)
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
		return fmt.Errorf("could not parse %q: %w", myURL, err)
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
		return fmt.Errorf("could not create folder to store html for debugging: %w", err)
	}
	f, err := os.Create("debug/" + n + ".html")
	if err != nil {
		return fmt.Errorf("unable to create file for %q: %w", myURL, err)
	}
	defer f.Close()
	_, err = f.WriteString(s)
	if err != nil {
		return fmt.Errorf("unable to write file for %q: %w", myURL, err)
	}
	err = f.Sync()
	if err != nil {
		return fmt.Errorf("unable to perform sync for %q: %w", myURL, err)
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
		return fmt.Errorf("unable to marshal search results: %w", err)
	}
	f, err := os.Create(fmt.Sprintf("searchResults/%v", n))
	if err != nil {
		return fmt.Errorf("unable to create file for search results: %w", err)
	}
	defer f.Close()
	_, err = f.Write(sr)
	if err != nil {
		return fmt.Errorf("unable to write search results to file: %w", err)
	}
	err = f.Sync()
	if err != nil {
		return fmt.Errorf("unable to perform sync for search results: %w", err)
	}
	return nil
}
