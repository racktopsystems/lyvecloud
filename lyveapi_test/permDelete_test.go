package lyveapi_test

import (
	"net/http"
	"testing"

	"github.com/racktopsystems/lyvecloud/lyveapi"
	"github.com/steinfletcher/apitest"
)

func TestPermissionsDelete(t *testing.T) {
	const pDoesNotExistJSONObj = `{
		"code": "PermissionNotFound",
		"message": "Permission was not found"
	}`

	const pDoesNotExistRespJSONObj = pDoesNotExistJSONObj

	var permissionDoesNotExist = apitest.NewMock().
		Delete(mockPermissionDelUri1).
		RespondWith().
		Body(pDoesNotExistJSONObj).
		Status(http.StatusNotFound).
		End()

	var permissionExists = apitest.NewMock().
		Delete(mockPermissionDelUri2).
		RespondWith().
		Status(http.StatusOK).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(permissionDoesNotExist).
		Handler(deletePermissionHandler()).
		Delete(mockPermissionDelUri1).
		Expect(t).
		Body(pDoesNotExistRespJSONObj).
		Status(http.StatusNotFound).
		End()

	apitest.New().
		Report(apitest.SequenceDiagram()).
		Mocks(permissionExists).
		Handler(deletePermissionHandler()).
		Delete(mockPermissionDelUri2).
		Expect(t).
		Status(http.StatusOK).
		End()
}

type mockDeletePermResp struct {
	err            *lyveapi.ApiCallFailedError
	httpStatusCode int
}

func deletePermissionHandler() *http.ServeMux {
	handler := http.NewServeMux()
	handler.HandleFunc(mockPermissionDelUri1, func(w http.ResponseWriter, r *http.Request) {
		errResponse := lyveapi.ApiCallFailedError{}
		if err := doDelete(r.URL.Path, &errResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Below we generate the response which ends-up being the Status Code
		// and the Body evaluated by the *apitest.APITest(s) established with
		// apitest.New() function in the Test functions above.
		w.WriteHeader(errResponse.HttpStatusCode())
		bytes := errResponse.JSON()
		_, err := w.Write(bytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	handler.HandleFunc(mockPermissionDelUri2, func(w http.ResponseWriter, r *http.Request) {
		response := mockDeletePermResp{}
		if err := doDelete(r.URL.Path, &response); err == nil {
			w.WriteHeader(response.httpStatusCode)
		} else {
			w.WriteHeader(http.StatusInternalServerError)

		}
		return
	})
	return handler

}

func doDelete(path string, response interface{}) error {
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	// This case matches an expected failure to delete permission due to
	// missing permission Id.
	case mockPermissionDelUri1:
		permissionId := path[len(mockPermissionsUri+"/"):]
		if err := client.DeletePermission(permissionId); err != nil {
			*response.(*lyveapi.ApiCallFailedError) =
				*err.(*lyveapi.ApiCallFailedError)
		}

	case mockPermissionDelUri2:
		permissionId := path[len(mockPermissionsUri+"/"):]
		if err := client.DeletePermission(permissionId); err != nil {
			return err
		}
		response.(*mockDeletePermResp).httpStatusCode = http.StatusOK
	}

	return nil
}
