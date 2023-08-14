package lyveapi_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/steinfletcher/apitest"
	"github.com/szaydel/lyvecloud/lyveapi"
)

var pathUnmatchedErr = errors.New("no path matched")
var unexpectedSuccessErr = errors.New("expected function to return a non-nil error")

const (
	mockApiEndpointUrl   = "https://localhost:8080/v2"
	mockAuthTokenUri     = "/auth/token"
	mockPermissionsUri   = "/permissions"
	mockOnePermissionUri = mockPermissionsUri + "/" + "mock-permission-1"
)

const createPermBadPolicyRespJSONObj = `{
	"code": "InternalError",
	"message": "The server encountered an internal error. Please retry the request."
}`

const onePermissionJSONObj = `{
	"id": "mock-permission-1",
	"name": "mock-all",
	"description": "Mock description",
	"type": "bucket-names",
	"readyState": true,
	"actions": "all-operations",
	"prefix": "mock-prefix",
	"buckets": [
		"mock-bucket"
	],
	"policy": "mock-policy"
}`
const pListJSONObj = `[{ 
	"name": "alpha",
	"id": "a1-b2-c3",
	"description": "something test", 
	"type": "all-buckets",
	"readyState": true,
	"createTime": "2023-08-11T19:15:00Z"
   },
   { 
	"name": "beta",
	"id": "c1-b2-a3",
	"description": "something test", 
	"type": "all-buckets",
	"readyState": true,
	"createTime": "2023-08-11T19:15:00Z"
   }]`

func TestPermissions(t *testing.T) {
	var permissionsListMock = apitest.NewMock().
		Get(mockApiEndpointUrl + mockPermissionsUri).
		RespondWith().
		Body(pListJSONObj).
		Status(http.StatusOK).
		End()

	var onePermissionMock = apitest.NewMock().
		Get(mockOnePermissionUri).
		RespondWith().
		Body(onePermissionJSONObj).
		Status(http.StatusOK).
		End()

	var permissionBadPolicyMock = apitest.NewMock().
		Post(mockPermissionsUri).
		RespondWith().
		Body(createPermBadPolicyRespJSONObj).
		Status(http.StatusInternalServerError).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(permissionsListMock).
		Handler(permissionsHandler()).
		Get(mockPermissionsUri).
		Expect(t).
		Body(pListJSONObj).
		Status(http.StatusOK).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(onePermissionMock).
		Handler(permissionsHandler()).
		Get(mockOnePermissionUri).
		Expect(t).
		Body(onePermissionJSONObj).
		Status(http.StatusOK).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(permissionBadPolicyMock).
		Handler(permissionsHandler()).
		Post(mockPermissionsUri).
		Expect(t).
		Body(createPermBadPolicyRespJSONObj).
		Status(http.StatusInternalServerError).
		End()
}

const tokenAcquisitionBadAuthJSONObj = `
{
	"code": "AuthenticationFailed ",
	"message": "Authentication failed."
}`

const tokenAcquisitionJSONObj = `
{ 
	"token": "this-is-a-mock-token",
	"expirationSec": "30"
}`

const tokenAcquisitionJSONReqObj = `
{ 
	"accountId": "mock-account",
	"accessKey": "mock-access-key",
	"secret": "mock-secret"
}`

const tokenValidJSONObj = `
{
	"expirationSec": "30"
}`

func TestToken(t *testing.T) {
	var tokenValidationMock = apitest.NewMock().
		Get(mockApiEndpointUrl + mockAuthTokenUri).
		RespondWith().
		Body(tokenValidJSONObj).
		Status(http.StatusOK).
		End()

	var tokenAcquisitionMock = apitest.NewMock().
		Post(mockApiEndpointUrl + mockAuthTokenUri).
		RespondWith().
		Body(tokenAcquisitionJSONObj).
		Status(http.StatusOK).
		End()

	var tokenAcquisitionBadAuthMock = apitest.NewMock().
		Post(mockApiEndpointUrl + mockAuthTokenUri).
		RespondWith().
		Body(tokenAcquisitionBadAuthJSONObj).
		Status(http.StatusForbidden).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(tokenValidationMock).
		Handler(tokenHandler()).
		Get(mockAuthTokenUri).
		Expect(t).
		Body(tokenValidJSONObj).
		Status(http.StatusOK).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(tokenAcquisitionMock).
		Handler(tokenHandler()).
		Post(mockAuthTokenUri).
		Body(tokenAcquisitionJSONReqObj).
		Expect(t).
		Body(tokenAcquisitionJSONObj).
		Status(http.StatusOK).
		End()

	// This is a bad authentication response test.
	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(tokenAcquisitionBadAuthMock).
		Handler(tokenHandler()).
		Post(mockAuthTokenUri).
		Body(tokenAcquisitionJSONReqObj).
		Expect(t).
		Body(tokenAcquisitionBadAuthJSONObj).
		Status(http.StatusForbidden).
		End()
}

func permissionsHandler() *http.ServeMux {
	var handler = http.NewServeMux()

	handler.HandleFunc(mockPermissionsUri, func(w http.ResponseWriter, r *http.Request) {
		var permission lyveapi.Permission
		var permissions lyveapi.PermissionList

		switch r.Method {
		case http.MethodGet:
			if err := httpGet(mockPermissionsUri, &permissions); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			bytes, _ := json.Marshal(permissions)
			_, err := w.Write(bytes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)

		case http.MethodPost:
			if err := permsBadHttpPost(
				mockPermissionsUri, &permission); err != nil {
				asErr := &lyveapi.ApiCallFailedError{}
				if errors.As(err, &asErr) {
					// The order here is important. First, call the WriteHeader
					// method to set the http.StatusInternalServerError
					// response code. Next, write the necessary JSON payload.
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(createPermBadPolicyRespJSONObj))
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
				return
			}

			bytes, _ := json.Marshal(permission)
			_, err := w.Write(bytes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	})

	handler.HandleFunc(mockOnePermissionUri, func(w http.ResponseWriter, r *http.Request) {
		var permission lyveapi.Permission
		if err := httpGet(mockOnePermissionUri, &permission); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bytes, _ := json.Marshal(permission)
		_, err := w.Write(bytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	return handler
}

func tokenHandler() *http.ServeMux {
	var handler = http.NewServeMux()
	handler.HandleFunc(mockAuthTokenUri, func(w http.ResponseWriter, r *http.Request) {
		var token lyveapi.Token
		switch r.Method {
		case http.MethodGet:
			if err := httpGet(mockAuthTokenUri, &token); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		case http.MethodPost:
			if err := httpPost(mockAuthTokenUri, &token); err != nil {
				if errors.Is(err, lyveapi.AuthenticationFailedErr) {
					// The order here is important. First, call the WriteHeader
					// method to set the http.StatusForbidden response code.
					// Next, write the necessary JSON payload.
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(tokenAcquisitionBadAuthJSONObj))
				} else {
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

func httpGet(path string, response interface{}) error {
	var err error
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockPermissionsUri:
		var p *lyveapi.PermissionList
		if p, err = client.ListPermissions(); err != nil {
			return err
		}
		pListPtr := response.(*lyveapi.PermissionList)
		*pListPtr = append(*pListPtr, *p...)
		return nil

	case mockOnePermissionUri:
		var p *lyveapi.Permission
		permissionId := mockOnePermissionUri[len("/permissions/"):]
		if p, err = client.GetPermission(permissionId); err != nil {
			return err
		}
		response.(*lyveapi.Permission).Id = p.Id
		response.(*lyveapi.Permission).Name = p.Name
		response.(*lyveapi.Permission).Description = p.Description
		response.(*lyveapi.Permission).Type = p.Type
		response.(*lyveapi.Permission).ReadyState = p.ReadyState
		response.(*lyveapi.Permission).Actions = p.Actions
		response.(*lyveapi.Permission).Prefix = p.Prefix
		response.(*lyveapi.Permission).Buckets = p.Buckets
		response.(*lyveapi.Permission).Policy = p.Policy
		return nil

	case mockAuthTokenUri:
		var t time.Time
		if t, err = client.TokenValidUntil(); err != nil {
		} else {
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

func httpPost(path string, response interface{}) error {
	var err error
	var token *lyveapi.Token
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockPermissionsUri:
		policy := `{"garbage": "ploicy"}`
		permission := &lyveapi.Permission{
			Name:        "mock-permission-1",
			Description: "Mock description",
			Type:        lyveapi.Policy,
			Prefix:      "pre",
			Policy:      policy,
			Actions:     lyveapi.AllOperations}
		if _, err := client.CreatePermission(permission); err != nil {
			return err
		}

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

func permsBadHttpPost(path string, response interface{}) error {
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockPermissionsUri:
		policy := `{"garbage": "ploicy"}`
		permission := &lyveapi.Permission{
			Name:        "mock-permission-1",
			Description: "Mock description",
			Type:        lyveapi.Policy,
			Prefix:      "pre",
			Policy:      policy,
			Actions:     lyveapi.AllOperations}
		if _, err := client.CreatePermission(permission); err != nil {
			return err
		} else {
			return unexpectedSuccessErr
		}
	}

	return pathUnmatchedErr
}
