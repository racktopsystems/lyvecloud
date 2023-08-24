package lyveapi_test

import "errors"

const (
	mockAuthTokenUri        = "/auth/token"
	mockPermissionIdAlpha   = "alpha-permission"
	mockPermission1         = "mock-permission-1"
	mockPermission2         = "mock-permission-2"
	mockPermissionIdBeta    = "beta-permission"
	mockApiEndpointUrl      = "https://localhost:8080/v2"
	mockPermissionsUri      = "/permissions"
	mockPermissionCreateUri = mockPermissionsUri + "/" + "mock-permission-1"
	mockPermissionUpdateUri = mockPermissionsUri + "/" + "mock-permission-2"
	mockPermissionDelUri1   = mockPermissionsUri + "/" + mockPermissionIdAlpha
	mockPermissionDelUri2   = mockPermissionsUri + "/" + mockPermissionIdBeta

	createPermBadPolicyRespJSONObj = `{
		"code": "InternalError",
		"message": "The server encountered an internal error. Please retry the request."
	}`
)

var pathUnmatchedErr = errors.New("no path matched")
var unexpectedSuccessErr = errors.New("expected function to return a non-nil error")
