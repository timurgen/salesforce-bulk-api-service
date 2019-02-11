package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Printf("Starting service on port %s", port)

	router := mux.NewRouter()
	router.HandleFunc("/{sObject}", FetchData).Methods("GET")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}

func FetchData(w http.ResponseWriter, r *http.Request) {
	username := os.Getenv("SALESFORCE_USERNAME")
	password := os.Getenv("SALESFORCE_PASSWORD") + os.Getenv("SALESFORCE_USER_TOKEN")
	params := mux.Vars(r)
	sObject := params["sObject"]
	var useSandbox = true

	if os.Getenv("SANDBOX") == "" {
		useSandbox = false
	}

	var api = CreateNew(username, password)
	err := api.LoginSoap(useSandbox)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Obtained session")

	objDescription, err := api.DescribeSObject(sObject)
	objFields := objDescription["fields"]
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	job, err := api.CreateJob(Query, sObject, "JSON")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Created job with id %s", job.Id)

	job.populateObjectFields(objFields.([]interface{}))

	err = api.AddBatchToJob(job)
	if err != nil {
		log.Printf("Error occurred while adding batch to job. Closing job %s", job.Id)
		api.CloseJob(job)
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	start := time.Now()
	for time.Now().Sub(start) < 120*time.Second {
		log.Printf("Checking status for job %s", job.Id)
		status, err := api.CheckJobStatus(job)
		log.Printf("Status: %s", status)
		if err != nil {
			log.Println(err)
		}
		if status == Completed {
			log.Printf("Job completed in %f seconds", time.Now().Sub(start).Seconds())
			break
		}
		time.Sleep(2 * time.Second)
	}

	result, err := api.GetJobResult(job)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Result size is %d elements", len(result))

	log.Printf("Closing job #%s", job.Id)
	err = api.CloseJob(job)
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("["))
	var first = true
	for _, item := range result {
		if first {
			first = false
		} else {
			w.Write([]byte(","))
		}
		jsonData, err := json.Marshal(item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonData)
	}
	w.Write([]byte("]"))
}
