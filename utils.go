package main

import (
	"fmt"
	"regexp"
	"strings"
)

func formatString(query string, args ...string) string{
	r := strings.NewReplacer(args...)
	res := fmt.Sprintf(r.Replace(query))
	return res
}

func extractInstanceFromUrl(url string) string {
	re := regexp.MustCompile("https://(.*?)\\.")
	match := re.FindStringSubmatch(url)
	return match[1]
}
