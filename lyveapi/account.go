package lyveapi

import (
	"encoding/json"
	"io"
	"net/http"
)

// CreateServiceAccount creates an account described by the createReq parameter
// returns a nil and an error if decoding of the response fails, otherwise a
// decoded object and nil error is returned.
func (client *Client) CreateServiceAccount(
	createReq *CreateServiceAcctReq) (*CreateServiceAcctResp, error) {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/service-accounts"
	token := client.token
	client.mtx.RUnlock()

	var err error
	var buf []byte
	var rdr io.ReadCloser

	if buf, err = json.Marshal(createReq); err != nil {
		return nil, err
	}

	if rdr, err = apiRequestAuthenticated(
		token, http.MethodPost, endpoint, buf); err != nil {
		return nil, err
	}

	defer rdr.Close()

	respBodyDecoder := json.NewDecoder(rdr)

	svcAcctResp := &CreateServiceAcctResp{}
	// This is where we fail when the v2/service-accounts/ request URI is
	// missing a trailing "/".
	if err := respBodyDecoder.Decode(svcAcctResp); err != nil {
		return nil, err
	}

	return svcAcctResp, nil
}

// ListServiceAccounts returns a list of service account details and an error.
// A successful request will result in service account listing and nil error,
// whereas a nil and an error is returned on failure.
func (client *Client) ListServiceAccounts() (*ServiceAcctList, error) {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/service-accounts"
	token := client.token
	client.mtx.RUnlock()

	var err error
	var rdr io.ReadCloser

	if rdr, err = apiRequestAuthenticated(
		token, http.MethodGet, endpoint, nil); err != nil {
		return nil, err
	}

	defer rdr.Close()

	svcAccts := &ServiceAcctList{}
	respBodyDecoder := json.NewDecoder(rdr)
	if err := respBodyDecoder.Decode(svcAccts); err != nil {
		return nil, err
	}

	return svcAccts, nil
}

// GetServiceAccount returns information about a Service Account if one is
// found. A successful request will result in a account details and a nil error,
// whereas a nil and an error is returned on failure.
func (client *Client) GetServiceAccount(svcAcctId string) (*ServiceAcct, error) {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/service-accounts"
	url := endpoint + "/" + svcAcctId
	token := client.token
	client.mtx.RUnlock()

	var err error
	var rdr io.ReadCloser

	if rdr, err = apiRequestAuthenticated(
		token, http.MethodGet, url, nil); err != nil {
		return nil, err
	}

	defer rdr.Close()

	var acctInfo = &ServiceAcct{}
	respBodyDecoder := json.NewDecoder(rdr)
	if err := respBodyDecoder.Decode(acctInfo); err != nil {
		return nil, err
	}
	return acctInfo, nil
}

func (client *Client) UpdateServiceAccount(
	svcAcctId string, changes ServiceAcctUpdateReq) error {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/service-accounts"
	url := endpoint + "/" + svcAcctId
	token := client.token
	client.mtx.RUnlock()

	var err error
	var rdr io.ReadCloser
	var data []byte

	if data, err = json.Marshal(changes); err != nil {

		return err
	}

	if rdr, err = apiRequestAuthenticated(
		token, http.MethodPut, url, data); err != nil {
		return err
	}

	// We may not have anything to read, but there may be a reader that we need
	// to close.
	if rdr != nil {
		rdr.Close()
	}

	// There should be no body if we got back a 200.
	return nil
}

// EnableServiceAccount enables an account associated with the given service
// account Id. A successful request will return a nil, whereas an error is
// returned if no such account could be found or some other error occurs.
func (client *Client) EnableServiceAccount(svcAcctId string) error {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/service-accounts"
	url := endpoint + "/" + svcAcctId + "/enabled"
	token := client.token
	client.mtx.RUnlock()

	var err error
	var rdr io.ReadCloser

	if rdr, err = apiRequestAuthenticated(token, http.MethodPut, url, nil); err != nil {
		return err
	}

	// We may not have anything to read, but there may be a reader that we need
	// to close.
	if rdr != nil {
		rdr.Close()
	}

	// There should be no body if we got back a 200.
	return nil
}

// DisableServiceAccount disables an account associated with the given service
// account Id. A successful request will return a nil, whereas an error is
// returned if no such account could be found or some other error occurs.
func (client *Client) DisableServiceAccount(svcAcctId string) error {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/service-accounts"
	url := endpoint + "/" + svcAcctId + "/enabled"
	token := client.token
	client.mtx.RUnlock()

	var err error
	var rdr io.ReadCloser

	if rdr, err = apiRequestAuthenticated(token, http.MethodDelete, url, nil); err != nil {
		return err
	}

	// We may not have anything to read, but there may be a reader that we need
	// to close.
	if rdr != nil {
		rdr.Close()
	}

	// There should be no body if we got back a 200.
	return nil
}

// DeleteServiceAccount deletes an account associated with the given service
// account Id. A successful request will return a nil, whereas an error is
// returned if no such account could be found.
func (client *Client) DeleteServiceAccount(svcAcctId string) error {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/service-accounts"
	url := endpoint + "/" + svcAcctId
	token := client.token
	client.mtx.RUnlock()

	var err error
	var rdr io.ReadCloser

	if rdr, err = apiRequestAuthenticated(token, http.MethodDelete, url, nil); err != nil {
		return err
	}

	// We may not have anything to read, but there may be a reader that we need
	// to close.
	if rdr != nil {
		rdr.Close()
	}

	// There should be no body if we got back a 200.
	return nil
}
