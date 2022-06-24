package main

import (
	"fmt"
	"os"
	"regexp"
)

// writeScript writes a passed in string of javascript to an individual file.
func writeScript(script, url string, fileNamer *regexp.Regexp) error {
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
func writeFile(scripts []string, fName string) error {
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
