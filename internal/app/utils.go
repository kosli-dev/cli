package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Contains checks if a string is contained in a string slice
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// DoPost sends an HTTP Post request to a URL and returns the response body and status code
func DoPost(jsonBody []byte, url string, apiToken string) (*HTTPResponse, error) {
	requestBody := bytes.NewBuffer(jsonBody)

	client := &http.Client{}

	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to create post request to %s : %v", url, err)
	}
	req.SetBasicAuth(apiToken, "unset")
	resp, err := client.Do(req)

	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to send post request to %s : %v", url, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{}, fmt.Errorf("failed to read response from post request to %s : %v", url, err)
	}

	return &HTTPResponse{
		Body:       string(body),
		StatusCode: resp.StatusCode,
	}, nil
}
