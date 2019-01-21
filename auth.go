package main

import "encoding/xml"

const LOGIN_XML_PAYLOAD = `<?xml version="1.0" encoding="utf-8" ?>

<env:Envelope xmlns:xsd="http://www.w3.org/2001/XMLSchema"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:env="http://schemas.xmlsoap.org/soap/envelope/">
  <env:Body>
    <n1:login xmlns:n1="urn:partner.soap.sforce.com">
      <n1:username>{username}</n1:username>
      <n1:password>{password}</n1:password>
    </n1:login>
  </env:Body>
</env:Envelope>`

const LOGIN_URL = `https://{env}.salesforce.com/services/Soap/u/{api_version}`

type SalesforceSoapFaultEnvelope struct {
	XMLName xml.Name
	Body    Body
}

type Body struct {
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

