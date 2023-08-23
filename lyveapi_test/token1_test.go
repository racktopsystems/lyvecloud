package lyveapi_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/steinfletcher/apitest"
	"github.com/szaydel/lyvecloud/lyveapi"
)

func TestToken(t *testing.T) {
	const mockTokenValid = `
	{
		"expirationSec": "30"
	}`

	const mockTokenAcquisitionBadAuthResp = `{
		"code": "AuthenticationFailed ",
		"message": "Authentication failed."
	}`

	const mockTokenAcquisitionResp = `{
		"token": "this-is-a-mock-token",
		"expirationSec": "30"
	}`

	const mockTokenAcquisitionReq = `{
		"accountId": "mock-account",
		"accessKey": "mock-access-key",
		"secret": "mock-secret"
	}`

	var tokenValidationMock = apitest.NewMock().
		Get(mockApiEndpointUrl + mockAuthTokenUri).
		RespondWith().
		Body(mockTokenValid).
		Status(http.StatusOK).
		End()

	var tokenAcquisitionMock = apitest.NewMock().
		Post(mockApiEndpointUrl + mockAuthTokenUri).
		RespondWith().
		Body(mockTokenAcquisitionResp).
		Status(http.StatusOK).
		End()

	var tokenAcquisitionBadAuthMock = apitest.NewMock().
		Post(mockApiEndpointUrl + mockAuthTokenUri).
		RespondWith().
		Body(mockTokenAcquisitionBadAuthResp).
		Status(http.StatusForbidden).
		End()

	// This is a token validation response test.
	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(tokenValidationMock).
		Handler(tokenHandler()).
		Get(mockAuthTokenUri).
		Expect(t).
		Body(mockTokenValid).
		Status(http.StatusOK).
		End()

	// This is a good authentication response test.
	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(tokenAcquisitionMock).
		Handler(tokenHandler()).
		Post(mockAuthTokenUri).
		Body(mockTokenAcquisitionReq).
		Expect(t).
		Body(mockTokenAcquisitionResp).
		Status(http.StatusOK).
		End()

	// This is a bad authentication response test.
	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(tokenAcquisitionBadAuthMock).
		Handler(tokenHandler()).
		Post(mockAuthTokenUri).
		Body(mockTokenAcquisitionReq).
		Expect(t).
		Body(mockTokenAcquisitionBadAuthResp).
		Status(http.StatusForbidden).
		End()
}

func tokenHandler() *http.ServeMux {
	var handler = http.NewServeMux()
	handler.HandleFunc(mockAuthTokenUri, func(w http.ResponseWriter, r *http.Request) {
		var token lyveapi.Token
		switch r.Method {
		case http.MethodGet:
			if err := authToken(mockAuthTokenUri, &token); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		case http.MethodPost:
			if err := doTokenPost(mockAuthTokenUri, &token); err != nil {
				// The order here is important. First, call the WriteHeader
				// method to set the http.StatusForbidden response code.
				// Next, write the necessary JSON payload.
				w.WriteHeader(err.(*lyveapi.ApiCallFailedError).HttpStatusCode())
				if _, err = w.Write(err.(*lyveapi.ApiCallFailedError).JSON()); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
		}

		bytes, _ := json.Marshal(token)
		_, err := w.Write(bytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	return handler
}

func authToken(path string, response interface{}) error {
	var err error
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockAuthTokenUri:
		var t time.Time
		if t, err = client.TokenValidUntil(); err == nil {
			remains := t.Sub(time.Now().
				Round(1_000_000_000)).
				Round(1_000_000_000).String()

			tokenPtr := response.(*lyveapi.Token)
			tokenPtr.ExpirationSec = remains[:len(remains)-1]
		}
		return err
	}

	return pathUnmatchedErr
}

func doTokenPost(path string, response interface{}) error {
	var err error
	var token *lyveapi.Token
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockAuthTokenUri:
		if token, err = lyveapi.Authenticate(&lyveapi.Credentials{}, mockApiEndpointUrl+path); err != nil {
			return err
		}
		response.(*lyveapi.Token).Token = token.Token
		response.(*lyveapi.Token).ExpirationSec = token.ExpirationSec
		return err
	}

	return pathUnmatchedErr
}
