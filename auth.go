package main

import "encoding/xml"

//Salesforce Slogi nSOAP request
const LOGIN_XML_PAYLOAD = `<?xml version="1.0" encoding="utf-8" ?>
<env:Envelope xmlns:xsd="http://www.w3.org/2001/XMLSchema"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:env="http://schemas.xmlsoap.org/soap/envelope/">
  <env:SalesforceLoginSoapResponseBody>
    <n1:login xmlns:n1="urn:partner.soap.sforce.com">
      <n1:username>{username}</n1:username>
      <n1:password>{password}</n1:password>
    </n1:login>
  </env:SalesforceLoginSoapResponseBody>
</env:Envelope>`

//Template for login URL
//{env} - which Salesforce environment to use login/test
//{api_version} - which API version to use
const LOGIN_URL = `https://{env}.salesforce.com/services/Soap/u/{api_version}`

type SalesforceLoginSoapResponse struct {
	XMLName xml.Name
	Body    SalesforceLoginSoapResponseBody
}

type SalesforceLoginSoapResponseBody struct {
	XMLName xml.Name
	Fault   faultResponse   `xml:"Fault"`
	Success successResponse `xml:"loginResponse"`
}

type faultResponse struct {
	XMLName     xml.Name
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
}

type successResponse struct {
	XMLName xml.Name
	Result  successResponseResult `xml:"result"`
}

type successResponseResult struct {
	XMLName           xml.Name
	MetadataServerUrl string `xml:"metadataServerUrl"`
	PasswordExpired   bool   `xml:"passwordExpired"`
	Sandbox           string `xml:"sandbox"`
	ServerUrl         string `xml:"serverUrl"`
	SessionId         string `xml:"sessionId"`
	UserId            string `xml:"userId"`
}

