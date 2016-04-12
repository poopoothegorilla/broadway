package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/namely/broadway/env"
	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/services"
	"github.com/namely/broadway/store"
	"github.com/namely/broadway/testutils"

	"github.com/stretchr/testify/assert"
)

var testToken = "BroadwayTestToken"

func makeRequest(req *http.Request, w *httptest.ResponseRecorder) {
	mem := store.New()

	server := New(mem).Handler()
	server.ServeHTTP(w, req)
}

func auth(req *http.Request) *http.Request {
	req.Header.Add("Authorization", "Bearer "+env.AuthBearerToken)
	return req
}

func TestServerNew(t *testing.T) {
	env.SlackToken = testToken

	mem := store.New()

	s := New(mem)
	assert.Equal(t, testToken, s.slackToken, "Expected server.slackToken to match existing ENV value")

	env.SlackToken = ""
	s = New(mem)
	assert.Equal(t, "", s.slackToken, "Expected server.slackToken to be empty string")

}

func TestAuthFailure(t *testing.T) {
	env.AuthBearerToken = "testtoken"
	w, server := helperSetupServer()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Bearer faketoken")
	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected POST / with wrong auth token to be 401")
}

func TestAuthSuccess(t *testing.T) {
	env.AuthBearerToken = "testtoken"
	w, server := helperSetupServer()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Bearer testtoken")
	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected POST / with correct auth token to be 200")
}

func TestAuthFailureHints(t *testing.T) {
	env.AuthBearerToken = "testtoken"
	w, server := helperSetupServer()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Bearer faketoken")
	server.ServeHTTP(w, req)
	assert.Contains(t, w.Body.String(), "Wrong")

	req, _ = http.NewRequest("GET", "/", nil)
	server.ServeHTTP(w, req)
	assert.Contains(t, w.Body.String(), "Missing")
}

func TestInstanceCreateWithValidAttributes(t *testing.T) {

	i := map[string]interface{}{
		"playbook_id": "helloplaybook",
		"id":          "TestInstanceCreateWithValidAttributes",
		"vars": map[string]string{
			"word": "gorilla",
		},
	}

	rbody := testutils.JSONFromMap(t, i)
	req, w := testutils.PostRequest(t, "/instances", rbody)
	req = auth(req)
	makeRequest(req, w)

	assert.Equal(t, http.StatusCreated, w.Code, "Response code should be 201")
}

func TestCreateInstanceWithInvalidAttributes(t *testing.T) {

	invalidRequests := map[string]map[string]interface{}{
		"playbook_id": {
			"id": "test",
			"vars": map[string]string{
				"version": "ok",
			},
		},
	}

	for _, i := range invalidRequests {
		rbody := testutils.JSONFromMap(t, i)
		req, w := testutils.PostRequest(t, "/instances", rbody)
		req = auth(req)
		makeRequest(req, w)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected POST /instances with wrong attributes to be 400")
	}
}

func TestGetInstanceWithValidPath(t *testing.T) {
	store := store.New()
	i := &instance.Instance{PlaybookID: "helloplaybook", ID: "TestGetInstanceWithValidPath"}
	service := services.NewInstanceService(store)
	_, err := service.Create(i)
	if err != nil {
		t.Log(err.Error())
	}

	req, w := testutils.GetRequest(t, "/instance/helloplaybook/TestGetInstanceWithValidPath")
	req = auth(req)
	makeRequest(req, w)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetInstanceWithInvalidPath(t *testing.T) {
	req, w := testutils.GetRequest(t, "/instance/vanished/TestGetInstanceWithInvalidPath")
	req = auth(req)
	makeRequest(req, w)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetInstancesWithFullPlaybook(t *testing.T) {
	testInstance1 := &instance.Instance{PlaybookID: "helloplaybook", ID: "TestGetInstancesWithFullPlaybook1"}
	testInstance2 := &instance.Instance{PlaybookID: "helloplaybook", ID: "TestGetInstancesWithFullPlaybook2"}
	service := services.NewInstanceService(store.New())
	_, err := service.Create(testInstance1)
	_, err = service.Create(testInstance2)
	if err != nil {
		t.Log(err.Error())
	}

	req, w := testutils.GetRequest(t, "/instances/helloplaybook")
	req = auth(req)
	makeRequest(req, w)

	assert.Equal(t, http.StatusOK, w.Code, "Response code should be 200 OK")
}

func TestGetStatusFailures(t *testing.T) {
	invalidRequests := []struct {
		method  string
		path    string
		errCode int
		errMsg  string
	}{
		{
			"GET",
			"/status/helloplaybook/TestGetStatusFailures",
			404,
			"Not Found",
		},
	}

	for _, i := range invalidRequests {
		req, w := testutils.GetRequest(t, i.path)
		req = auth(req)
		makeRequest(req, w)

		assert.Equal(t, i.errCode, w.Code)

		var errorResponse map[string]string

		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.Nil(t, err)
		assert.Contains(t, errorResponse["error"], i.errMsg)
	}

}
func TestGetStatusWithGoodPath(t *testing.T) {
	testInstance1 := &instance.Instance{
		PlaybookID: "helloplaybook",
		ID:         "TestGetStatusWithGoodPath",
		Status:     instance.StatusDeployed}
	is := services.NewInstanceService(store.New())
	_, err := is.Create(testInstance1)
	if err != nil {
		t.Error(err)
		return
	}
	req, w := testutils.GetRequest(t, "/status/helloplaybook/TestGetStatusWithGoodPath")
	req = auth(req)
	makeRequest(req, w)

	assert.Equal(t, http.StatusOK, w.Code)

	var statusResponse map[string]string

	err = json.Unmarshal(w.Body.Bytes(), &statusResponse)
	assert.Nil(t, err)
	assert.Contains(t, statusResponse["status"], "deployed")
}

func helperSetupServer() (*httptest.ResponseRecorder, http.Handler) {
	w := httptest.NewRecorder()
	mem := store.New()
	server := New(mem).Handler()
	return w, server
}

func TestGetCommand400(t *testing.T) {
	w, server := helperSetupServer()
	req, err := http.NewRequest("GET", "/command", nil)
	if err != nil {
		t.Fatal(err)
	}

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected GET /command to be 400")
}

func TestGetCommand200(t *testing.T) {
	w, server := helperSetupServer()
	req, _ := http.NewRequest("GET", "/command?ssl_check=1", nil)

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected GET /command?ssl_check=1 to be 200")
}
func TestPostCommandMissingToken(t *testing.T) {
	env.SlackToken = testToken
	w, server := helperSetupServer()
	formBytes := bytes.NewBufferString("not a form")
	req, _ := http.NewRequest("POST", "/command", formBytes)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected POST /command with bad body to be 401")
}
func TestPostCommandWrongToken(t *testing.T) {
	env.SlackToken = testToken
	w, server := helperSetupServer()
	req, _ := http.NewRequest("POST", "/command", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form := url.Values{}
	form.Set("token", "wrongtoken")
	req.PostForm = form

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected POST /command with wrong token to be 401")
}
func TestPostCommandHelp(t *testing.T) {
	env.SlackToken = testToken
	w, server := helperSetupServer()
	req, _ := http.NewRequest("POST", "/command", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form := url.Values{}
	form.Set("token", testToken)
	form.Set("command", "/broadway")
	form.Set("text", "help")
	req.PostForm = form

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected /broadway help to be 200")
	assert.Contains(t, w.Body.String(), "/broadway", "Expected help message to contain /broadway")
}

func TestSlackCommandSetvar(t *testing.T) {
	env.SlackToken = testToken
	w, server := helperSetupServer()
	req, _ := http.NewRequest("POST", "/command", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form := url.Values{}
	form.Set("token", testToken)
	form.Set("command", "/broadway")
	form.Set("text", "setvar boing bar var1=val1")
	req.PostForm = form

	i := &instance.Instance{PlaybookID: "boing", ID: "bar", Vars: map[string]string{"var1": "val2"}}
	is := services.NewInstanceService(store.New())
	_, err := is.Create(i)
	if err != nil {
		t.Log(err)
	}
	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected slack command to be 200")
}

func TestPostCommandDeployBad(t *testing.T) {
	env.SlackToken = testToken
	w, server := helperSetupServer()
	req, _ := http.NewRequest("POST", "/command", nil)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	form := url.Values{}
	form.Set("token", testToken)
	form.Set("command", "/broadway")
	form.Set("text", "deploy foo")
	req.PostForm = form

	server.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Expected /broadway deploy foo to be 200")
	assert.Contains(t, w.Body.String(), "/broadway deploy myPlaybookID myInstanceID", "Expected help message to contain /broadway")
}

func TestDeployMissing(t *testing.T) {
	mem := store.New()
	w := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/deploy/missingPlaybook/missingInstance", nil)
	assert.Nil(t, err)
	req = auth(req)

	server := New(mem).Handler()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResponse map[string]string
	log.Println(w.Body.String())
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.Nil(t, err)
	assert.Contains(t, errorResponse["error"], "Not Found")
}
