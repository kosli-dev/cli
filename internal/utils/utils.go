package utils

import (
	"encoding/json"
	"io/ioutil"
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

// LoadFileContent loads file content
func LoadFileContent(filepath string) (string, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(data), err
}

// IsJSON checks if a string is in JSON format
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}
