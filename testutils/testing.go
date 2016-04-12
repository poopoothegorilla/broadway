package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// JSONFromMap creates a json representation from a map
func JSONFromMap(t *testing.T, data map[string]interface{}) []byte {
	rbody, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
		return []byte{}
	}
	return rbody
}

// PostRequest createa a post request for sending json
func PostRequest(t *testing.T, route string, data []byte) (*http.Request, *httptest.ResponseRecorder) {
	return buildRequest(t, "POST", route, data)
}

// GetRequest creates a get request
func GetRequest(t *testing.T, route string) (*http.Request, *httptest.ResponseRecorder) {
	return buildRequest(t, "GET", route, []byte{})
}

// DeleteRequest creates a delete request
func DeleteRequest(t *testing.T, route string) (*http.Request, *httptest.ResponseRecorder) {
	return buildRequest(t, "DELETE", route, []byte{})
}

func buildRequest(t *testing.T, action string, route string, data []byte) (*http.Request, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest(action, route, bytes.NewBuffer(data))
	if err != nil {
		t.Error(err)
		return req, w
	}
	req.Header.Add("Content-Type", "application/json")
	return req, w
}
