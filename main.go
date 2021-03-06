package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"
)

const DATE_LAYOUT = "2006-01-02T15:04:05.000-0700"

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
	if os.Getenv("DEBUG") != "" {
		f, err := os.Create(fmt.Sprintf("profile-%d.prof", time.Now().Unix()))
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	log.Printf("Serving request: %v", r)

	username := os.Getenv("SALESFORCE_USERNAME")
	password := os.Getenv("SALESFORCE_PASSWORD") + os.Getenv("SALESFORCE_USER_TOKEN")
	params := mux.Vars(r)
	sObject := params["sObject"]

	value := r.URL.Query()
	since := value.Get("since")
	var sinceTime time.Time

	if since != "" {
		t, err := time.Parse(DATE_LAYOUT, since)
		if err != nil {
			log.Println(err)
		} else {
			sinceTime = t
		}
	}

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

	err = api.AddBatchToJob(job, sinceTime)
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
	log.Println("Fetching Job results")
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
	log.Println("Sending response back to client")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("["))
	var first = true
	for idx, item := range result {
		if first {
			first = false
		} else {
			w.Write([]byte(","))
		}
		mappedItem := item.(map[string]interface{})
		switch updDate := mappedItem["LastModifiedDate"].(type) {
		case float64:
			mappedItem["_updated"] = time.Unix(int64(updDate)/1000, 0).Format(DATE_LAYOUT)
		case string:
			mappedItem["_updated"] = updDate
		}

		jsonData, err := json.Marshal(mappedItem)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonData)
		jsonData = nil
		item = nil
		if idx%10000 == 0 {
			debug.FreeOSMemory()
		}
	}
	w.Write([]byte("]"))
	log.Println("Request completed")

	if os.Getenv("DEBUG") != "" {
		log.Println("Some memory stats...")
		f, err := os.Create(fmt.Sprintf("profile-%d.mprof", time.Now().Unix()))
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		log.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
		log.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
		log.Printf("\tSys = %v MiB", m.Sys/1024/1024)
		log.Printf("\tNumGC = %v\n", m.NumGC)

		log.Println("Forcing GC...")
		debug.FreeOSMemory()

		runtime.ReadMemStats(&m)
		log.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
		log.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
		log.Printf("\tSys = %v MiB", m.Sys/1024/1024)
		log.Printf("\tNumGC = %v\n", m.NumGC)
	}

}
