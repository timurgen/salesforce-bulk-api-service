package main

import (
	"fmt"
	"regexp"
	"strings"
)

//formatString format string using named parameters
func formatString(query string, args ...string) string {
	r := strings.NewReplacer(args...)
	res := fmt.Sprintf(r.Replace(query))
	return res
}

//extractInstanceFromUrl extract Salesforce instance name from Salesforce url
func extractInstanceFromUrl(url string) string {
	re := regexp.MustCompile("https://(.*?)\\.")
	match := re.FindStringSubmatch(url)
	return match[1]
}

func stringInArray(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
