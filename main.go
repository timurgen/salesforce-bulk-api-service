package main

import "fmt"

func main() {
	var api = CreateNew("","")
	err := api.loginSoap(false)

	if err != nil {
		fmt.Printf("error: %v", err)
	}

	api.CreateJob(Query, "Contact", "JSON")
}
