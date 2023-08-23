package lyveapi_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/steinfletcher/apitest"
	"github.com/szaydel/lyvecloud/lyveapi"
)

func TestNewClient(t *testing.T) {
	const mockAuthTokenRespBody = `{ 
		"token": "this-is-a-mock-token",
		"expirationSec": "864000"
	}`

	var tokenValidationMock = apitest.NewMock().
		Post(mockAuthTokenUri).
		RespondWith().
		Body(mockAuthTokenRespBody).
		Status(http.StatusOK).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(tokenValidationMock).
		Handler(tokenHandler1()).
		Post(mockAuthTokenUri).
		Expect(t).
		Body(mockAuthTokenRespBody).
		Status(http.StatusOK).
		End()
}

func tokenHandler1() *http.ServeMux {
	var handler = http.NewServeMux()
	handler.HandleFunc(mockAuthTokenUri, func(w http.ResponseWriter, r *http.Request) {
		var token lyveapi.Token
		switch r.Method {
		// case http.MethodGet:
		// 	if err := authToken(mockAuthTokenUri, &token); err != nil {
		// 		w.WriteHeader(http.StatusInternalServerError)
		// 		return
		// 	}
		case http.MethodPost:
			creds := lyveapi.Credentials{
				AccountId: "mock-account",
				AccessKey: "mock-key",
				Secret:    "mock-secret",
			}
			if err := doNewClientToken(mockAuthTokenUri, &creds, &token); err != nil {
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

func doNewClientToken(path string, creds *lyveapi.Credentials, response interface{}) error {
	var err error
	var client *lyveapi.Client
	// client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockAuthTokenUri:
		if client, err = lyveapi.NewClient(creds, mockApiEndpointUrl); err != nil {
			return err
		}
		// if token, err = lyveapi.Authenticate(&lyveapi.Credentials{}, mockApiEndpointUrl+path); err != nil {
		// 	return err
		// }

		seconds := client.ExpiresAfter().Sub(time.Now().Round(1_000_000_000)).Milliseconds() / 1000

		expiresInString := strconv.FormatInt(seconds, 10)

		response.(*lyveapi.Token).Token = client.Token()
		response.(*lyveapi.Token).ExpirationSec = expiresInString
		return err
	}

	return pathUnmatchedErr
}
