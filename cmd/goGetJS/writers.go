package main

import (
	"fmt"
	"os"
	"regexp"
)

func writeScripts(script, url string) error {
	r := regexp.MustCompile(`[\w-]+(\.js)?$`)
	fileName := r.FindString(url)
	scriptByte := []byte(script)
	if err := os.WriteFile("data/"+fileName, scriptByte, 0644); err != nil {
		return err
	}
	return nil
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
