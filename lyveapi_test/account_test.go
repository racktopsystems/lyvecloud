package lyveapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/racktopsystems/lyvecloud/lyveapi"
	"github.com/steinfletcher/apitest"
)

const mockSvcAcctUpdateAcctId = "037c16bc-1409-4997-a8d7-523b985e32d9"
const mockOneSvcAccountUri = mockSvcAcctsUri + "/" + mockSvcAcctUpdateAcctId
const mockSvcAccountEnableDisableUri = mockOneSvcAccountUri + "/" + "enabled"

func TestSvcAccounts(t *testing.T) {
	const sGetSvcAcctRespBody = `{
      "id": "c79ab3e6-f3b4-4265-81bd-db3eee8ce213",
      "name": "alphatest3",
      "description": "alphatest3 service account",
      "enabled": true,
      "expirationDate": "",
      "readyState": false,
      "permissions": [
        "4440bab8-6525-4760-bb26-22bedfd5195a"
      ]
    }`
	const sCreateSvcAcctRespBody = `{
      "id": "c79ab3e6-f3b4-4265-81bd-db3eee8ce213",
      "accessKey": "EXHJ4WFC3UE3891B",
      "secret": "KVKIBHEPUYKXVKYGS4I8C021DFZVNHL3",
      "expirationDate": ""
    }`
	const sListSvcAcctsRespBody = `[
    {
      "description": "alpha test service account",
      "enabled": true,
	  "expirationDate": "",
      "id": "b99ca0e6-d6ed-4c1c-9687-fa90fec287e1",
      "name": "alpha",
      "readyState": false
    },
    {
      "description": "beta test service account",
      "enabled": true,
	  "expirationDate": "",
      "id": "420573f9-953d-4e91-ae79-fd81a91dcbf7",
      "name": "beta",
      "readyState": true
    }
  ]`

	var svcAcctCreateSuccessMock = apitest.NewMock().
		Post(mockSvcAcctsUri).
		RespondWith().
		Body(sCreateSvcAcctRespBody).
		Status(http.StatusOK).
		End()

	var svcAcctDeleteSuccessMock = apitest.NewMock().
		Delete(mockOneSvcAccountUri).
		RespondWith().
		Status(http.StatusOK).
		End()

	var svcAcctDisableSuccessMock = apitest.NewMock().
		Delete(mockSvcAccountEnableDisableUri).
		RespondWith().
		Status(http.StatusOK).
		End()

	var svcAcctEnableSuccessMock = apitest.NewMock().
		Put(mockSvcAccountEnableDisableUri).
		RespondWith().
		Status(http.StatusOK).
		End()

	var svcAcctUpdateSuccessMock = apitest.NewMock().
		Put(mockSvcAcctsUri).
		RespondWith().
		Status(http.StatusOK).
		End()

	var svcAcctsListMock = apitest.NewMock().
		Get(mockSvcAcctsUri).
		RespondWith().
		Body(sListSvcAcctsRespBody).
		Status(http.StatusOK).
		End()

	var svcAcctGetOneMock = apitest.NewMock().
		Get(mockOneSvcAccountUri).
		RespondWith().
		Body(sGetSvcAcctRespBody).
		Status(http.StatusOK).
		End()

	apitest.New("get service account").
		Report(apitest.SequenceDiagram()).
		Mocks(svcAcctGetOneMock).
		Handler(svcAccountsHandler()).
		Get(mockOneSvcAccountUri).
		Expect(t).
		Body(sGetSvcAcctRespBody).
		Status(http.StatusOK).
		End()

	apitest.New("create service account").
		Report(apitest.SequenceDiagram()).
		Mocks(svcAcctCreateSuccessMock).
		Handler(svcAccountsHandler()).
		Post(mockSvcAcctsUri).
		Expect(t).
		Body(sCreateSvcAcctRespBody).
		Status(http.StatusOK).
		End()

	apitest.New("delete service account").
		Report(apitest.SequenceDiagram()).
		Mocks(svcAcctDeleteSuccessMock).
		Handler(svcAccountsHandler()).
		Delete(mockOneSvcAccountUri).
		Expect(t).
		Status(http.StatusOK).
		End()

	apitest.New("disable service account").
		Report(apitest.SequenceDiagram()).
		Mocks(svcAcctDisableSuccessMock).
		Handler(svcAccountsHandler()).
		Delete(mockSvcAccountEnableDisableUri).
		Expect(t).
		Status(http.StatusOK).
		End()

	apitest.New("enable service account").
		Report(apitest.SequenceDiagram()).
		Mocks(svcAcctEnableSuccessMock).
		Handler(svcAccountsHandler()).
		Put(mockSvcAccountEnableDisableUri).
		Expect(t).
		Status(http.StatusOK).
		End()

	apitest.New("updates to service account").
		Report(apitest.SequenceDiagram()).
		Mocks(svcAcctUpdateSuccessMock).
		Handler(svcAccountsHandler()).
		Put(mockOneSvcAccountUri).
		Expect(t).
		Status(http.StatusOK).
		End()

	apitest.New("list service accounts").
		Report(apitest.SequenceDiagram()).
		Mocks(svcAcctsListMock).
		Handler(svcAccountsHandler()).
		Get(mockSvcAcctsUri).
		Expect(t).
		Body(sListSvcAcctsRespBody).
		Status(http.StatusOK).
		End()
}

func svcAccountsHandler() *http.ServeMux {
	var handler = http.NewServeMux()

	handler.HandleFunc(mockSvcAccountEnableDisableUri, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete: // Disable the service account
			if err := invokeClientToDisableServiceAccount(mockSvcAccountEnableDisableUri); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)

		case http.MethodPut: // Enable the service account
			if err := invokeClientToEnableServiceAccount(mockSvcAccountEnableDisableUri); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	})

	handler.HandleFunc("/service-accounts/", func(w http.ResponseWriter, r *http.Request) {
		var svcAcct lyveapi.ServiceAcct

		switch r.Method {
		case http.MethodDelete: // Delete the service account
			if err := invokeClientToDeleteServiceAccount(mockOneSvcAccountUri); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodPut:
			req := &lyveapi.ServiceAcct{
				Name:        "new-test-name",
				Description: "This is a test account",
			}

			if err := invokeClientToUpdateServiceAccount(
				mockOneSvcAccountUri, req, &svcAcct); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)

		case http.MethodGet:
			if err := invokeClientToGetServiceAccount(
				mockOneSvcAccountUri, &svcAcct); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			} else {
				bytes, _ := json.Marshal(&svcAcct)
				_, err := w.Write(bytes)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			}
		}
	})

	handler.HandleFunc("/service-accounts", func(w http.ResponseWriter, r *http.Request) {
		var svcAcctsList []lyveapi.ServiceAcct
		var svcAcctCreateReq lyveapi.CreateServiceAcctReq
		var svcAcctCreateResp lyveapi.CreateServiceAcctResp

		switch r.Method {
		case http.MethodGet:
			if err := invokeClientToListServiceAccounts(mockSvcAcctsUri, &svcAcctsList); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			bytes, _ := json.Marshal(svcAcctsList)
			_, err := w.Write(bytes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)

		case http.MethodPost:
			svcAcctCreateReq = lyveapi.CreateServiceAcctReq{
				Name:        "alphatest",
				Description: "alphatest description",
				Permissions: []string{"fake-permission-id"},
			}
			if err := invokeClientToCreateServiceAccount(
				mockSvcAcctsUri, &svcAcctCreateReq, &svcAcctCreateResp); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			} else {
				bytes, _ := json.Marshal(&svcAcctCreateResp)
				_, err := w.Write(bytes)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			}

			w.WriteHeader(http.StatusOK)
		}

	})

	return handler
}

func invokeClientToDeleteServiceAccount(path string) error {
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockOneSvcAccountUri:
		tokens := strings.Split(path, "/")
		svcAcctId := tokens[2]
		if err := client.DeleteServiceAccount(svcAcctId); err != nil {
			return err
		} else {
			return nil
		}
	default:
		return fmt.Errorf("no matches for given API path: %q", path)
	}
}

func invokeClientToDisableServiceAccount(path string) error {
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockSvcAccountEnableDisableUri:
		tokens := strings.Split(path, "/")
		svcAcctId := tokens[2]
		if err := client.DisableServiceAccount(svcAcctId); err != nil {
			return err
		} else {
			return nil
		}
	default:
		return fmt.Errorf("no matches for given API path: %q", path)
	}
}

func invokeClientToEnableServiceAccount(path string) error {
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockSvcAccountEnableDisableUri:
		tokens := strings.Split(path, "/")
		svcAcctId := tokens[2]
		if err := client.EnableServiceAccount(svcAcctId); err != nil {
			return err
		} else {
			return nil
		}
	default:
		return fmt.Errorf("no matches for given API path: %q", path)
	}
}

func invokeClientToUpdateServiceAccount(
	path string, updates *lyveapi.ServiceAcct, response interface{}) error {
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockOneSvcAccountUri:
		tokens := strings.Split(path, "/")
		svcAcctId := tokens[2]
		if err := client.UpdateServiceAccount(svcAcctId, updates); err != nil {
			return err
		} else {
			return nil
		}
	default:
		return fmt.Errorf("no matches for given API path: %q", path)
	}
}

func invokeClientToListServiceAccounts(
	path string,
	response interface{},
) error {
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockSvcAcctsUri:
		if svcAccts, err := client.ListServiceAccounts(); err != nil {
			return err
		} else {
			var respSvcAcctsList *[]lyveapi.ServiceAcct
			respSvcAcctsList = response.(*[]lyveapi.ServiceAcct)
			for _, v := range *svcAccts {
				*respSvcAcctsList = append(*respSvcAcctsList, v)
			}
			return nil
		}
	default:
		return fmt.Errorf("no matches for given API path: %q", path)
	}
}

func invokeClientToCreateServiceAccount(
	path string,
	createReq *lyveapi.CreateServiceAcctReq,
	resp interface{}) error {
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockSvcAcctsUri:
		if createResp, err := client.CreateServiceAccount(createReq); err != nil {
			return err
		} else {
			resp.(*lyveapi.CreateServiceAcctResp).Id = createResp.Id
			resp.(*lyveapi.CreateServiceAcctResp).AccessKey = createResp.AccessKey
			resp.(*lyveapi.CreateServiceAcctResp).Secret = createResp.Secret
			resp.(*lyveapi.CreateServiceAcctResp).ExpirationDate = createResp.ExpirationDate
			return nil
		}
	default:
		return fmt.Errorf("no matches for given API path: %q", path)
	}
}

func invokeClientToGetServiceAccount(path string, resp interface{}) error {
	var client = &lyveapi.Client{}
	client.SetApiURL(mockApiEndpointUrl)

	switch path {
	case mockOneSvcAccountUri:
		tokens := strings.Split(path, "/")
		svcAcctId := tokens[2]
		if getResp, err := client.GetServiceAccount(svcAcctId); err != nil {
			return err
		} else {
			resp.(*lyveapi.ServiceAcct).Id = getResp.Id
			resp.(*lyveapi.ServiceAcct).Name = getResp.Name
			resp.(*lyveapi.ServiceAcct).Description = getResp.Description
			resp.(*lyveapi.ServiceAcct).Enabled = getResp.Enabled
			resp.(*lyveapi.ServiceAcct).ReadyState = getResp.ReadyState
			for _, p := range getResp.Permissions {
				resp.(*lyveapi.ServiceAcct).Permissions =
					append(resp.(*lyveapi.ServiceAcct).Permissions, p)
			}
			return nil
		}
	default:
		return fmt.Errorf("no matches for given API path: %q", path)
	}
}
