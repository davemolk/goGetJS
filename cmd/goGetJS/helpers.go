package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
)

// assertErrorToNil is a simple helper function for error handling.
func (app *application) assertErrorToNil(err error) {
	if err != nil {
		app.errorLog.Fatal(err)
	}
}

// readInputFile converts the contents of an input text file to a string slice.
func (app *application) readInputFile(n string) ([]string, error) {
	var lines []string
	f, err := os.Open(n)
	if err != nil {
		return lines, fmt.Errorf("open input file error: %w", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// getInput checks if the user has supplied a url via stdin.
// If no url is found, goGetJS will exit (getInput is only called
// in the event that no url has been supplied via flag.)
func (app *application) getInput() error {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("stdin read error: %w", err)
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			app.config.url = s.Text()
		}
	}
	if app.config.url == "" {
		return errors.New("must provide a url")
	}
	return nil
}

// getBaseURL takes the url from the user and returns the base url.
func (app *application) getBaseURL(myUrl string) (string, error) {
	u, err := url.Parse(myUrl)
	if err != nil {
		return "", fmt.Errorf("unable to parse url: %w", err)
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
		if err != nil {
			app.errorLog.Fatal(err)
		}
		app.query = query
	} else {
		app.query = app.config.term
	}
}
