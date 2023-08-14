package lyveapi

import (
	"encoding/json"
	"io"
	"net/http"
)

// CreatePermission creates a new permission with the specified parameters.
// A nil and an error are returned upon failure.
func (client *Client) CreatePermission(createReq *Permission) (*Permission, error) {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/permissions"
	token := client.token
	client.mtx.RUnlock()

	var buf []byte
	var err error
	var rdr io.ReadCloser

	// If the permission type is "policy", we must have a policy object
	// associated with the permission.
	if createReq.IsPolicyPermission() && createReq.Policy == "" {
		return nil, PolicyMissingErr
	}

	if buf, err = json.Marshal(createReq); err != nil {
		return nil, err
	}

	if rdr, err = apiRequestAuthenticated(
		token, http.MethodPost, endpoint, buf); err != nil {
		return nil, err
	}

	defer rdr.Close()

	var permission = &Permission{}
	respBodyDecoder := json.NewDecoder(rdr)
	if err := respBodyDecoder.Decode(permission); err != nil {
		return nil, err
	}
	return permission, nil
}

// ListPermissions produces a list of Permission structs that are part of this
// account. A nil and an error are returned upon failure.
func (client *Client) ListPermissions() (*PermissionList, error) {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/permissions"
	token := client.token
	client.mtx.RUnlock()

	var err error
	var rdr io.ReadCloser

	if rdr, err = apiRequestAuthenticated(
		token, http.MethodGet, endpoint, nil); err != nil {
		return nil, err
	}

	defer rdr.Close()

	var permsList = &PermissionList{}
	respBodyDecoder := json.NewDecoder(rdr)
	if err := respBodyDecoder.Decode(permsList); err != nil {
		return nil, err
	}
	return permsList, nil
}

// GetPermission retrieves the permission associated with the specified
// permission id if one was found. If a permission for the specified id is not
// found An nil and an error will be returned.
func (client *Client) GetPermission(permissionId string) (*Permission, error) {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/permissions"
	url := endpoint + "/" + permissionId
	token := client.token
	client.mtx.RUnlock()

	var err error
	var rdr io.ReadCloser

	if rdr, err = apiRequestAuthenticated(
		token, http.MethodGet, url, nil); err != nil {
		return nil, err
	}

	defer rdr.Close()

	var permission = &Permission{}
	respBodyDecoder := json.NewDecoder(rdr)
	if err := respBodyDecoder.Decode(permission); err != nil {
		return nil, err
	}
	return permission, nil

}

// DeletePermission deletes a permission associated with the given permission
// Id. A successful request will return a nil, whereas an error is
// returned if no such permission could be found.
func (client *Client) DeletePermission(permissionId string) error {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/permissions"
	url := endpoint + "/" + permissionId
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

// UpdatePermission updates a permission associated with the given permission
// Id. A successful request will return a nil, whereas an error is returned if
// the request failed.
func (client *Client) UpdatePermission(
	permissionId string, updateReq *Permission) error {
	client.mtx.RLock()
	endpoint := client.apiUrl + "/permissions"
	url := endpoint + "/" + permissionId
	token := client.token
	client.mtx.RUnlock()

	var buf []byte
	var err error
	var rdr io.ReadCloser

	if buf, err = json.Marshal(updateReq); err != nil {
		return err
	}

	if rdr, err = apiRequestAuthenticated(
		token, http.MethodPut, url, buf); err != nil {
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
