package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const JOB_REQUEST_PAYLOAD = `<?xml version="1.0" encoding="UTF-8"?>
<jobInfo xmlns="http://www.force.com/2009/06/asyncapi/dataload">
    <operation>{operation}</operation>
    <object>{sObject}</object>
    <contentType>{contentType}</contentType>
</jobInfo>`

const SERVICE_URL = `https://{instance}.salesforce.com/services/async/{api_version}/job`

type Operation string

const (
	Insert   Operation = "insert"
	Update   Operation = "update"
	Delete   Operation = "delete"
	Upsert   Operation = "upsert"
	Query    Operation = "query"
	QueryAll Operation = "queryAll"
)

type Job struct {
	XMLName                 xml.Name
	Id                      string `xml:"id"`
	Operation               string `xml:"operation"`
	Object                  string `xml:"object"`
	CreatedById             string `xml:"createdById"`
	CreatedDate             string `xml:"createdDate"`
	State                   string `xml:"state"`
	NumberBatchesQueued     int    `xml:"numberBatchesQueued"`
	NumberBatchesInProgress int    `xml:"numberBatchesInProgress"`
	NumberBatchesCompleted  int    `xml:"numberBatchesCompleted"`
	NumberBatchesFailed     int    `xml:"numberBatchesFailed"`
	NumberBatchesTotal      int    `xml:"numberBatchesTotal"`
	NumberRecordsProcessed  int    `xml:"numberRecordsProcessed"`
	NumberRetries           int    `xml:"numberRetries"`
	ApiVersion              string `xml:"apiVersion"`
	NumberRecordsFailed     int    `xml:"numberRecordsFailed"`
	TotalProcessingTime     int    `xml:"totalProcessingTime"`
	ApiActiveProcessingTime int    `xml:"apiActiveProcessingTime"`
	ApexProcessingTimeint   int    `xml:"apexProcessingTime"`
	Batch                   []batch
	ObjectFields            []string
}

type batch struct {
	XMLName xml.Name
	Batchinfo batchinfo `xml:"batchInfo"`
}

type batchinfo struct {
	XMLName xml.Name
	Id string `xml:"id"`
	JobId string `xml:"jobId"`
	State string `xml:"state"`
	CreatedDate string `xml:"createdDate"`
	SystemModstamp string `xml:"systemModstamp"`
	NumberRecordsProcessed int `xml:"numberRecordsProcessed"`
	NumberRecordsFailed int `xml:"numberRecordsFailed"`
	TotalProcessingTime int `xml:"totalProcessingTime"`
	ApiActiveProcessingTime int `xml:"apiActiveProcessingTime"`
	ApexProcessingTime int `xml:"apexProcessingTime"`

}

type Api struct {
	apiVersion string
	instance   string
	username   string
	userId     string
	password   string
	sessionId  string
	serverUrl  string
	client     *http.Client
}

func CreateNew(username string, password string) *Api {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	return &Api{username: username, password: password, apiVersion: "44.0", client: httpClient}
}

//loginSoap authenticates using credentials provided  in Api object
func (api *Api) loginSoap(sandbox bool) error {
	var env = "test"
	if !sandbox {
		env = "login"
	}
	var loginUrl = formatString(LOGIN_URL, "{env}", env, "{api_version}", api.apiVersion)
	var reqPayload = formatString(LOGIN_XML_PAYLOAD, "{username}", api.username, "{password}", api.password)

	req, err := http.NewRequest("POST", loginUrl, strings.NewReader(reqPayload))

	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "text/xml; charset=UTF-8")
	req.Header.Add("SOAPAction", "login")
	resp, err := api.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		//Salesforce sends 500 HTTP and SOAP body with explanation
		faultResponse := &SalesforceSoapFaultEnvelope{}
		bodyBytes, innerErr := ioutil.ReadAll(resp.Body)
		if innerErr != nil {
			return innerErr
		}
		innerErr = xml.Unmarshal(bodyBytes, faultResponse)
		if innerErr != nil {
			return innerErr
		}
		return errors.New(fmt.Sprintf("Error: %s caused %s",
			faultResponse.Body.Fault.FaultCode, faultResponse.Body.Fault.FaultString))
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	successRes := &SalesforceSoapFaultEnvelope{}

	err = xml.Unmarshal(bodyBytes, successRes)
	if err != nil {
		return err
	}

	api.serverUrl = successRes.Body.Success.Result.ServerUrl
	api.sessionId = successRes.Body.Success.Result.SessionId
	api.userId = successRes.Body.Success.Result.UserId
	api.instance = extractInstanceFromUrl(api.serverUrl)

	return nil
}

func (api *Api) CreateJob(operation Operation, sObject string, contentType string) (*Job, error) {
	var result = &Job{}
	var reqPayload = formatString(JOB_REQUEST_PAYLOAD,
		"{operation}", string(operation), "{sObject}", sObject, "{contentType}", contentType)
	var serviceUrl = formatString(SERVICE_URL, "{instance}", api.instance, "{api_version}", api.apiVersion)

	req, err := http.NewRequest("POST", serviceUrl, strings.NewReader(reqPayload))
	if err != nil {
		return result, err
	}

	req.Header.Add("X-SFDC-Session", api.sessionId)
	req.Header.Add("Content-Type", "application/xml; charset=UTF-8")
	resp, err := api.client.Do(req)

	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	err = xml.Unmarshal(bodyBytes, result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (api *Api) AddBatchToJob(job *Job)  error {
	var reqPayload = fmt.Sprintf("SELECT Id, Name from %s", job.Object)
	var jobUrl = formatString(SERVICE_URL,
		"{instance}", api.instance, "{api_version}", api.apiVersion) + "/" + job.Id + "/batch"

	req, err := http.NewRequest("POST", jobUrl, strings.NewReader(reqPayload))
	if err != nil {
		return err
	}

	req.Header.Add("X-SFDC-Session", api.sessionId)
	req.Header.Add("Content-Type", "application/json") //fixme define type dynamically
	resp, err := api.client.Do(req)
	if err != nil {
		return  err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

}
