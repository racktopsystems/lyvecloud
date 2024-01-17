package lyveapi_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/racktopsystems/lyvecloud/lyveapi"
	"github.com/steinfletcher/apitest"
)

const mockApiAuthenticationUrl = mockApiEndpointUrl + mockAuthTokenUri

func TestNewClient(t *testing.T) {
	const mockCredentialBody = `{"accountId": "mock-account","accessKey": "alpha-beta","secret": "gamma-delta"}`

	const mockAuthTokenRespBody = `{ 
		"token": "alpha-beta-gamma-delta",
		"expirationSec": "864000"
	}`

	var tokenValidationMock = apitest.NewMock().
		Post(mockApiAuthenticationUrl).
		RespondWith().
		Body(mockAuthTokenRespBody).
		Status(http.StatusOK).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(tokenValidationMock).
		Handler(authenticationHandler()).
		Post(mockApiAuthenticationUrl).
		JSON(mockCredentialBody).
		Expect(t).
		Body(mockAuthTokenRespBody).
		Status(http.StatusOK).
		End()
}

func authenticationHandler() *http.ServeMux {
	var handler = http.NewServeMux()
	handler.HandleFunc("/v2/auth/token", func(w http.ResponseWriter, r *http.Request) {
		var token lyveapi.Token
		switch r.Method {
		case http.MethodPost:
			var credBytes []byte
			if b, err := io.ReadAll(r.Body); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				credBytes = b
			}

			creds := lyveapi.Credentials{}
			if err := json.Unmarshal(credBytes, &creds); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if err := invokeNewClientFactory(mockApiEndpointUrl, &creds, &token); err != nil {
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

func invokeNewClientFactory(path string, creds *lyveapi.Credentials, response interface{}) error {
	var err error
	var client *lyveapi.Client

	switch path {
	case mockApiEndpointUrl:
		if client, err = lyveapi.NewClient(creds, path); err != nil {
			return err
		}

		seconds := client.ExpiresAfter().Sub(time.Now().Round(1_000_000_000)).
			Milliseconds() / 1000

		expiresInString := strconv.FormatInt(seconds, 10)

		response.(*lyveapi.Token).Token = client.Token()
		response.(*lyveapi.Token).ExpirationSec = expiresInString
		return err
	}

	return pathUnmatchedErr
}
