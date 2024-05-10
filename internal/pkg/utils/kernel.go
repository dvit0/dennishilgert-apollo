package utils

import (
	"encoding/json"
	"net/url"
)

// SerializeJson serializes a map to a URL-encoded JSON string.
func SerializeJson(object map[string]interface{}) (string, error) {
	jsonString, err := json.Marshal(object)
	if err != nil {
		return "", err
	}
	encodedJson := url.QueryEscape(string(jsonString))
	return encodedJson, nil
}

// DeserializeJson deserializes a URL-encoded JSON string to a map.
func DeserializeJson(serializedJson string) (map[string]interface{}, error) {
	decodedJson, err := url.QueryUnescape(serializedJson)
	if err != nil {
		return nil, err
	}
	var object map[string]interface{}
	err = json.Unmarshal([]byte(decodedJson), &object)
	if err != nil {
		return nil, err
	}
	return object, nil
}
