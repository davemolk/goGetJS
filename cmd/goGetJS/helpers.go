package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"regexp"
)

// assertErrorToNilf is a simple helper function for error handling.
func (app *application) assertErrorToNilf(msg string, err error) {
	if err != nil {
		app.errorLog.Fatalf(msg, err)
	}
}

// readInputFile converts the contents of an input text file to a string slice.
func (app *application) readInputFile(n string) ([]string, error) {
	var lines []string
	f, err := os.Open(n)
	if err != nil {
		return lines, fmt.Errorf("could not open input file: %v", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// getBaseURL takes in the url from the user and returns the base url.
func (app *application) getBaseURL(myUrl string) (string, error) {
	u, err := url.Parse(myUrl)
	if err != nil {
		return "", fmt.Errorf("unable to parse input url %s: %v", myUrl, err)
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

// getQuery checks whether or not the user has used a term flag, a
// regex flag, or a terms flag. If any of these has been submitted, 
// the respective input is stored in the query field of the site struct.
func (app *application) getQuery() {
	if len(app.config.regex) > 0 {
		re := regexp.MustCompile(app.config.regex)
		app.query = re
	} else if len(app.config.terms) > 0 {
		query, err := app.readInputFile(app.config.terms)
		app.assertErrorToNilf("unable to get input for query %v", err)
		app.query = query
	} else {
		app.query = app.config.term
	}
}
