package main

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
)

// assertErrorToNilf is a simple helper function for error handling.
func (app *application) assertErrorToNilf(msg string, err error) {
	if err != nil {
		app.errorLog.Fatalf(msg, err)
	}
}

// readLines converts the contents of an input text file to a string slice.
func (app *application) readLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// getBaseURL takes in the url from the user and returns the base url (just in case...)
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
