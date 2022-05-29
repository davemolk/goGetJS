package main

import (
	"log"
	"net/http"
	"strings"
)

func getAbsURL(res *http.Response) string {
	base := *res.Request.URL
	abs := base.String()
	return strings.TrimSuffix(abs, "/")
}

func assertErrorToNilf(msg string, err error) {
	if err != nil {
		log.Fatalf(msg, err)
	}
}