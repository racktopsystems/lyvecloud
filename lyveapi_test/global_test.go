package lyveapi_test

import "errors"

const (
	mockAuthTokenUri      = "/auth/token"
	mockPermissionId1     = "alpha-permission"
	mockPermissionId2     = "beta-permission"
	mockApiEndpointUrl    = "https://localhost:8080/v2"
	mockPermissionsUri    = "/permissions"
	mockOnePermissionUri  = mockPermissionsUri + "/" + "mock-permission-1"
	mockPermissionDelUri1 = mockPermissionsUri + "/" + mockPermissionId1
	mockPermissionDelUri2 = mockPermissionsUri + "/" + mockPermissionId2

	createPermBadPolicyRespJSONObj = `{
		"code": "InternalError",
		"message": "The server encountered an internal error. Please retry the request."
	}`
)

var pathUnmatchedErr = errors.New("no path matched")
var unexpectedSuccessErr = errors.New("expected function to return a non-nil error")
