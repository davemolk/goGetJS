package main

import (
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
		return fmt.Errorf("cannot write script %q: %v", fName, err)
	}
	return nil
}

// writeFile takes a string slice of src and writes the contents to a text file.
func (app *application) writeFile(scripts []string, fName string) error {
	f, err := os.Create(fName)
	if err != nil {
		return fmt.Errorf("could not create %q: %v", fName, err)
	}
	defer f.Close()
	for _, v := range scripts {
		if _, err := fmt.Fprintln(f, v); err != nil {
			return fmt.Errorf("unable to write %q in %q: %v", v, fName, err)
		}
	}
	return nil
}

// writePage takes the html of a page (as a string) and the url and writes 
// the contents to a text file.
func (app *application) writePage(s, myURL string) error {
	var n string // for naming purposes
	u, err := url.Parse(myURL)
	if err != nil {
		return fmt.Errorf("could not parse %q: %v", myURL, err)
	}
	switch {
	case u.Path != "":
		re := regexp.MustCompile(`[-\w\?\=]+/?$`)
		n = re.FindString(u.Path)
		n =	strings.TrimSuffix(n, "/")
	default:
		n = u.Host
	}
	f, err := os.Create("pages/" + n + ".html")
	if err != nil {
		return fmt.Errorf("unable to create file for %q: %v", s, err)
	}
	defer f.Close()
	_, err = f.WriteString(s)
	if err != nil {
		return fmt.Errorf("unable to write file for %q: %v", s, err)
	}
	f.Sync()
	return nil
}