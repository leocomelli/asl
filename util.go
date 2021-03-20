package main

import "regexp"

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

// Snake converts text to use slash as a separator
func Snake(txt string) string {
	snake := matchFirstCap.ReplaceAllString(txt, "${1}-${2}")
	return matchAllCap.ReplaceAllString(snake, "${1}-${2}")
}
