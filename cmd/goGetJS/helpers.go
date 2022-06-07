package main

import (
	"bufio"
	"io"
	"log"
)

// assertErrorToNilf is a simple helper function for error handling.
func assertErrorToNilf(msg string, err error) {
	if err != nil {
		log.Fatalf(msg, err)
	}
}

// readLines converts the contents of an input text file to a string slice.
func readLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
