package lyveapi

import (
	"encoding/json"
	"errors"
)

var (
	InvalidTokenErr         = errors.New("token presented to the API is invalid")
	AuthenticationFailedErr = errors.New("authentication was unsuccessful; check supplied credentials")
	PermissionExistsErr     = errors.New("permission name is already taken")
	PolicyMissingErr        = errors.New("permission is missing required policy JSON document")
	PermissionNoExistErr    = errors.New("permission does not exist")
)

var errorCodesToErrors = map[string]error{
	"InvalidToken":         InvalidTokenErr,
	"AuthenticationFailed": AuthenticationFailedErr,
	// We are seemingly getting a trailing space in auth failure responses.
	"AuthenticationFailed ":       AuthenticationFailedErr,
	"PermissionNameAlreadyExists": PermissionExistsErr,
	"PermissionNotFound":          PermissionNoExistErr,
}

type requestFailedResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ApiCallFailedError struct {
	apiResp        *requestFailedResp
	httpStatusCode int
}

func (e *ApiCallFailedError) Code() string {
	return e.apiResp.Code
}

func (e *ApiCallFailedError) Message() string {
	return e.apiResp.Message
}

func (e *ApiCallFailedError) HttpStatusCode() int {
	return e.httpStatusCode
}

func (e *ApiCallFailedError) Error() string {
	var code, message string
	if e.apiResp.Message != "" {
		message = e.apiResp.Message
	} else {
		message = "no additional information given"
	}

	if e.apiResp.Code != "" {
		code = e.apiResp.Code
	} else {
		code = "unknown"
	}

	return "request failed: " + message + " " + "(" + code + ")"
}

func (e *ApiCallFailedError) JSON() []byte {
	b, _ := json.Marshal(e.apiResp)
	return b
}
