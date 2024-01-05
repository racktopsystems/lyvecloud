package lyveapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// decodeFailedApiResponse takes a response object from the API and converts it
// into a more user-friendly native representation. It returns an exported
// error type, which has methods for accessing the status code from the API and
// the message.
func decodeFailedApiResponse(resp *http.Response) error {
	body := &bytes.Buffer{}
	tRdr := io.TeeReader(resp.Body, body)
	decoder := json.NewDecoder(tRdr)
	respPayload := &requestFailedResp{}

	// If we can successfully decode the failure payload we encode the response
	// from the API as a ApiCallFailedError and return that to the caller.
	// Otherwise we have to try to deal with JSON error parsing and return
	// whatever we can figure out to the call to enable troubleshooting.
	if err := decoder.Decode(respPayload); err == nil {
		err = &ApiCallFailedError{
			apiResp:        respPayload,
			httpStatusCode: resp.StatusCode,
		}
		return err
	}

	// Dealing with an error in JSON parsing. This is due to the API not always
	// adhering to the specified contract and responding with HTML instead of
	// JSON-serialized data.
	// We expect that this is HTML and starts with a '<head>' and has multiple
	// lines. We want to eliminate the '\r\n' bits and present this garbage as a
	// single line, mostly for debug-ability.
	var errMsg string
	if bodySlc, err := io.ReadAll(body); err != nil {
		return errors.New("Failed to read contents of the response body")
	} else {
		if len(bodySlc) > 0 {
			if bodySlc[0] == '<' {
				var c int
				b := make([]byte, len(bodySlc))
				for _, v := range bodySlc {
					if v == '\n' || v == '\r' {
						continue
					}
					b[c] = v
					c++
				}
				bodySlc = b
				errMsg = string(bodySlc)
			}
		}
	}

	// At this point we may have some garbage, but let's return that anyway. :(
	return errors.New("dubious response from the API: " + errMsg)
}

// apiRequestAuthenticated packages up requests to the API without attempting
// to authenticate first. A valid token is required to complete requests
// successfully.
func apiRequestAuthenticated(
	token, method, url string, payload []byte) (io.ReadCloser, error) {
	headers := map[string][]string{
		"Accept": {
			"application/json",
		},
		"Authorization": {
			"Bearer " + token,
		},
	}

	var data *bytes.Buffer
	var req *http.Request
	var resp *http.Response
	var err error
	// If we are supplying a payload, we have to additionally set the
	// "Content-Type" header.
	if method != http.MethodGet {
		headers["Content-Type"] = []string{
			"application/json",
		}
		data = bytes.NewBuffer(payload)
		req, err = http.NewRequest(method, url, data)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	req.Header = headers
	client := &http.Client{}
	resp, err = client.Do(req)

	// If the error is non-nil, we should not expect a usable body. Therefore
	// we do not attempt to close the body at this point.
	// We should not expect err != nil if the status code from the API is
	// anything other than 200.
	if err != nil {
		return nil, err
	}

	// DEBUGGING:
	// buf, _ := io.ReadAll(resp.Body)
	// body := bytes.NewBuffer(buf)
	// resp.Body = io.NopCloser(body)

	// log.Printf("DEBUG (response body): %s", body.String())

	// Check response from the API and if resp.StatusCode != http.StatusOK, we
	// are going to have access to the error object which we should return to
	// the caller.
	// If the response is not http.StatusOK, look for an error response object.
	if resp.StatusCode != http.StatusOK {
		if resp.Body == nil {
			return nil, errors.New("non-200 response did not come with any reason for failure")
		}

		// Re-enable bits below for additional debugging
		// respBody := make([]byte, 4096)
		// resp.Body.Read(respBody)
		// log.Print("DEBUG: url: ", url)
		// log.Print("DEBUG: response body: ", string(respBody))
		// We need to be sure to close the body, since we are not going to
		// return it to the caller in this error path.
		defer resp.Body.Close()

		// If we encountered an issue decoding the body, return nil along with
		// the error surfaced during decoding. Otherwise return nil along with
		// decoded error as ApiCallFailedError.
		if resp.ContentLength > 0 {
			return nil, decodeFailedApiResponse(resp)
		} else {
			return nil, errors.New("Unexpected empty body in the response")
		}
	}

	// Handle http.StatusOK response next.
	return resp.Body, nil
}

// Authenticate attempts to authenticate against the API and returns a token
// upon successful authentication. The API will expire this token after 24
// hours. This expiration period appears to be fixed but Lyve Cloud may change
// it at any time.
func Authenticate(ctx context.Context, credentials *Credentials, authEndpointUrl string) (*Token, error) {
	var data *bytes.Buffer

	headers := map[string][]string{
		"Content-Type": {
			"application/json",
		},
		"Accept": {
			"application/json",
		},
	}

	if buf, err := json.Marshal(credentials); err != nil {
		return nil, err
	} else {
		data = bytes.NewBuffer(buf)
	}

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, authEndpointUrl+"/auth/token", data)
	req.Header = headers

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBodyDecoder := json.NewDecoder(resp.Body)
	authTok := &Token{}

	if resp.StatusCode != 200 {
		return nil, decodeFailedApiResponse(resp)
	}

	if err := respBodyDecoder.Decode(authTok); err != nil {
		return nil, err
	}

	return authTok, nil
}
