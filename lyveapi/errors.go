package lyveapi

import (
	"encoding/json"
	"errors"
)

var (
	InvalidTokenErrMsg         = "token presented to the API is invalid"
	ExpiredTokenErrMsg         = "token presented to the API is already expired"
	AuthenticationFailedErrMsg = "authentication was unsuccessful; check supplied credentials"
	PermissionExistsErrMsg     = "permission name is already taken"
	PolicyMissingErrMsg        = "permission is missing required policy JSON document"
	PermissionNoExistErrMsg    = "permission does not exist"

	// These errors are not something we get back from the API and convert
	// into an ApiCallFailedError. These are internal, indicating improper
	// usage or some other fault condition.
	PolicyMissingErr = errors.New(PermissionExistsErrMsg)
)

var errorCodesToErrors = map[string]string{
	"ExpiredToken":         ExpiredTokenErrMsg,
	"InvalidToken":         InvalidTokenErrMsg,
	"AuthenticationFailed": AuthenticationFailedErrMsg,
	// We are seemingly getting a trailing space in auth failure responses.
	"AuthenticationFailed ":       AuthenticationFailedErrMsg,
	"PermissionNameAlreadyExists": PermissionExistsErrMsg,
	"PermissionNotFound":          PermissionNoExistErrMsg,
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
	if v, ok := errorCodesToErrors[e.apiResp.Code]; ok {
		message = v
	} else if e.apiResp.Message != "" {
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
