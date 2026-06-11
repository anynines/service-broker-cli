package sbcli

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type SBClient struct {
	Credentials
}

func (s *SBClient) SetCredentials(c Credentials) {
	s.Host = c.Host
	s.Username = c.Username
	s.Password = c.Password
	s.SkipSslValidation = c.SkipSslValidation
}

func (s *SBClient) isHttps() bool {
	return strings.HasPrefix(s.Host, "https")
}

func (s *SBClient) Catalog() (*Catalog, error) {
	result, _, _, err := s.getResultFromBroker("v2/catalog", "GET", "{}")
	if err != nil {
		return nil, err
	}

	var c = new(Catalog)
	err = json.Unmarshal(result, &c)
	if err != nil {
		return nil, err
	}
	return c, err
}

func (s *SBClient) TestConnection() error {
	if os.Getenv("SB_TRACE") == "ON" {
		fmt.Println("")
		fmt.Printf("\tTest host: %s\n", s.Host)
	}

	timeout := 15
	val, present := os.LookupEnv("SB_TIMEOUT")
	if present == true {
		timeout, _ = strconv.Atoi(val)
	}

	client := http.Client{Timeout: time.Duration(timeout) * time.Second}

	if s.isHttps() {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Get(s.Host)

	if err != nil {
		if os.Getenv("SB_TRACE") == "ON" {
			fmt.Println("")
			fmt.Printf("\tError: %s\n", err.Error())
		}
		return err
	}

	if os.Getenv("SB_TRACE") == "ON" {
		fmt.Println("")
		fmt.Print("\tStatus: OK\n")
	}
	defer resp.Body.Close()
	return nil
}

/* func (s *SBClient) LastState(instanceId string) (*LastState, error) {
	result, _, _, err := s.getResultFromBroker(fmt.Sprintf("v2/service_instances/%s/last_operation", instanceId), "GET", "{}")
	if err != nil {
		return nil, err
	}

	var l = new(LastState)
	err = json.Unmarshal(result, &l)
	if err != nil {
		return nil, err
	}
	return l, err
} */

func (s *SBClient) Instances() (*Instances, error) {
	var instanceResources []InstanceResource
	next_url := "v2/instances?results_per_page=5000"
	next_url_exists := true
	for next_url_exists == true {
		var response InstancesResponse
		result, _, _, err := s.getResultFromBroker(next_url, "GET", "{}")
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(result, &response)
		if err != nil {
			return nil, err
		}

		instanceResources = append(instanceResources, response.Resources...)

		if string(response.NextURL) == "" {
			next_url_exists = false
		}
		if next_url_exists == true {
			next_url = response.NextURL
		}
	}

	var i = new(Instances)
	i.Resources = instanceResources
	return i, nil
}

func (s *SBClient) Instance(instanceId string) (*InstanceResource, error) {
	result, _, _, err := s.getResultFromBroker(fmt.Sprintf("instances/%s", instanceId), "GET", "{}")
	if err != nil {
		return nil, err
	}

	var i = new(InstanceResource)
	err = json.Unmarshal(result, &i)
	if err != nil {
		return nil, err
	}
	return i, err
}

// doRequest performs the raw HTTP request and returns the response body, status code,
// status string, and any transport-level error. It does NOT parse broker error payloads.
func (s *SBClient) doRequest(url string, method string, jsonStr string) (result []byte, statusCode int, status string, err error) {
	statusCode = 0
	status = ""
	result = nil
	body := strings.NewReader(jsonStr)
	target := fmt.Sprintf("%s/%s", s.Host, url)

	if os.Getenv("SB_TRACE") == "ON" {
		fmt.Println("")
		fmt.Printf("\tRequest to: %s\n", target)
		fmt.Printf("\tMethod:     %s\n", method)
		fmt.Printf("\tBody:\n\t%s\n", jsonStr)
	}

	timeout := 15
	val, present := os.LookupEnv("SB_TIMEOUT")
	if present == true {
		timeout, _ = strconv.Atoi(val)
	}

	client := http.Client{Timeout: time.Duration(timeout) * time.Second}

	if s.isHttps() {
		if os.Getenv("SB_TRACE") == "ON" {
			fmt.Println("\tHTTPS:      true")
		}

		var tlsConfig tls.Config

		if s.SkipSslValidation {
			tlsConfig.InsecureSkipVerify = true
			if os.Getenv("SB_TRACE") == "ON" {
				fmt.Println("\tSkip SSL Verification: true ")
			}
		}

		t := &http.Transport{
			TLSClientConfig: &tlsConfig,
		}
		client.Transport = t
	}

	req, err := http.NewRequest(method, target, body)
	if err != nil {
		return
	}
	if s.Username != "" {
		req.SetBasicAuth(s.Username, s.Password)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		if os.Getenv("SB_TRACE") == "ON" {
			fmt.Printf("\tError:\n\t%s\n", err.Error())
		}
		return
	}

	defer resp.Body.Close()

	status = resp.Status
	statusCode = resp.StatusCode

	result, err = ioutil.ReadAll(resp.Body)
	if os.Getenv("SB_TRACE") == "ON" {
		fmt.Println("\tResult:")
		fmt.Printf("\tStatus: %d/%s\n", resp.StatusCode, resp.Status)
		fmt.Printf("\tBody:\n\t%s\n", string(result))
	}
	return
}

// getResultFromBroker performs a request and additionally parses broker error payloads
// from the response body, returning an error if the broker reported one.
func (s *SBClient) getResultFromBroker(url string, method string, jsonStr string) (bytes []byte, statusCode int, status string, err error) {
	bytes, statusCode, status, err = s.doRequest(url, method, jsonStr)
	if err != nil {
		return
	}

	var sbError = new(SBError)
	tempErr := json.Unmarshal(bytes, &sbError)
	if (sbError != nil && (sbError.Error != "" || sbError.Description != "")) || tempErr != nil {
		err = errors.New(fmt.Sprintf("%s / %s", sbError.Description, sbError.Error))
		return
	}

	return
}

// LastOperation polls GET /v2/service_instances/:id/last_operation and returns the result.
// It uses doRequest directly so that an informational "description" field in the response
// is not mistakenly treated as a broker error.
func (s *SBClient) LastOperation(instanceID string, operation string) (*LastOperationResponse, int, error) {
	path := fmt.Sprintf("v2/service_instances/%s/last_operation", instanceID)
	if operation != "" {
		path += "?operation=" + operation
	}
	bytes, statusCode, _, err := s.doRequest(path, "GET", "{}")
	if err != nil {
		return nil, statusCode, err
	}
	var resp LastOperationResponse
	if err = json.Unmarshal(bytes, &resp); err != nil {
		return nil, statusCode, err
	}
	return &resp, statusCode, nil
}

func (s *SBClient) Deprovision(data *BindPayload, instanceID string) (int, string, error) {
	bytes, statusCode, status, err := s.getResultFromBroker(fmt.Sprintf("v2/service_instances/%s?service_id=%s&plan_id=%s&accepts_incomplete=true", instanceID, data.ServiceID, data.PlanID), "DELETE", "{}")
	if err != nil {
		return statusCode, "", err
	}

	if statusCode >= 200 && statusCode <= 202 {
		var resp ProvisionResponse
		json.Unmarshal(bytes, &resp)
		return statusCode, resp.Operation, nil
	}

	return statusCode, "", errors.New(fmt.Sprintf("Deprovision failure code: %d/%s", statusCode, status))
}

func (s *SBClient) UpdateService(data *UpdatePayload, instanceID string) (int, string, error) {
	payloadBytes, _ := json.Marshal(data)

	resultBytes, statusCode, status, err := s.getResultFromBroker(fmt.Sprintf("v2/service_instances/%s?accepts_incomplete=true", instanceID), "PATCH", string(payloadBytes))
	if err != nil {
		return statusCode, "", err
	}

	if statusCode >= 200 && statusCode <= 202 {
		var resp ProvisionResponse
		json.Unmarshal(resultBytes, &resp)
		return statusCode, resp.Operation, nil
	}

	return statusCode, "", errors.New(fmt.Sprintf("Update failure code: %d/%s", statusCode, status))
}

func (s *SBClient) Provision(data *ProvisonPayload, instanceID string) (int, string, error) {
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return 0, "", err
	}

	resultBytes, statusCode, status, err := s.getResultFromBroker(fmt.Sprintf("v2/service_instances/%s?accepts_incomplete=true", instanceID), "PUT", string(payloadBytes))
	if err != nil {
		return statusCode, "", err
	}

	if statusCode >= 200 && statusCode <= 202 {
		var resp ProvisionResponse
		json.Unmarshal(resultBytes, &resp)
		return statusCode, resp.Operation, nil
	}

	return statusCode, "", errors.New(fmt.Sprintf("Provision failure code: %d/%s", statusCode, status))
}

func (s *SBClient) Bind(data *BindPayload, instanceID string, bindID string) (int, string, string, error) {
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return 0, "", "", err
	}

	resultBytes, statusCode, status, err := s.doRequest(fmt.Sprintf("v2/service_instances/%s/service_bindings/%s?accepts_incomplete=true", instanceID, bindID), "PUT", string(payloadBytes))
	if err != nil {
		return statusCode, "", "", err
	}

	// Parse broker error (but not on 202 where body contains operation, not an error)
	if statusCode != 202 {
		var sbError SBError
		if json.Unmarshal(resultBytes, &sbError) == nil && (sbError.Error != "" || sbError.Description != "") {
			return statusCode, "", "", errors.New(fmt.Sprintf("%s / %s", sbError.Description, sbError.Error))
		}
	}

	if statusCode >= 200 && statusCode <= 202 {
		var resp ProvisionResponse
		json.Unmarshal(resultBytes, &resp)
		return statusCode, resp.Operation, string(resultBytes), nil
	}

	return statusCode, "", "", errors.New(fmt.Sprintf("Bind failure code: %d/%s", statusCode, status))
}

func (s *SBClient) UnBind(data *BindPayload, instanceID string, bindID string) (int, string, error) {
	resultBytes, statusCode, status, err := s.doRequest(fmt.Sprintf("v2/service_instances/%s/service_bindings/%s?service_id=%s&plan_id=%s&accepts_incomplete=true", instanceID, bindID, data.ServiceID, data.PlanID), "DELETE", "{}")
	if err != nil {
		return statusCode, "", err
	}

	if statusCode >= 200 && statusCode <= 202 {
		var resp ProvisionResponse
		json.Unmarshal(resultBytes, &resp)
		return statusCode, resp.Operation, nil
	}

	// 410 Gone means already deleted
	if statusCode == 410 {
		return statusCode, "", nil
	}

	return statusCode, "", errors.New(fmt.Sprintf("Unbind failure code: %d/%s", statusCode, status))
}

// creates the Servicebroker client, in later version the user credentials should be read out of a file
func NewSBClient(cred ...*Credentials) *SBClient {
	var sb SBClient

	if len(cred) == 0 {
		conf := LoadConfig()
		sb.Host = conf.Host
		sb.Password = conf.Password
		sb.Username = conf.Username
		sb.SkipSslValidation = conf.SkipSslValidation
	} else {
		sb.Host = cred[0].Host
		sb.Username = cred[0].Username
		sb.Password = cred[0].Password
		sb.SkipSslValidation = cred[0].SkipSslValidation
	}

	return &sb
}
