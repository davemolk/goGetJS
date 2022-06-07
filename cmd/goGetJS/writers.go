package main

import (
	"fmt"
	"os"
	"regexp"
)

// writeScript writes a passed in string of javascript to an individual file.
func writeScript(script, url string, fileNamer *regexp.Regexp) error {
	fileName := fileNamer.FindString(url)
	scriptByte := []byte(script)
	if err := os.WriteFile("data/"+fileName, scriptByte, 0644); err != nil {
		return err
	}
	return nil
}

// writeFile takes a string slice of src and writes the contents to a text file.
func writeFile(scripts []string, fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, v := range scripts {
		if _, err := fmt.Fprintln(f, v); err != nil {
			return err
		}
	}
	return nil
}
