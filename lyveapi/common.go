package lyveapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func decodeFailedApiResponse(resp *http.Response) error {
	body := &bytes.Buffer{}
	tRdr := io.TeeReader(resp.Body, body)
	decoder := json.NewDecoder(tRdr)
	respPayload := &requestFailedResp{}

	// If we can successfully decode the failure payload, we lookup the status
	// code in the status codes to errors map. If a matching error is found, it
	// is returned to the caller. Otherwise we encode the response from the API
	// as a ApiCallFailedError and return that to the caller.
	if err := decoder.Decode(respPayload); err == nil {
		if mappedErr, ok := errorCodesToErrors[respPayload.Code]; ok {
			err = mappedErr
		} else {
			err = &ApiCallFailedError{
				apiResp: respPayload,
			}
		}
		return err
	}

	// Dealing with an error in JSON parsing. This is due to the API not always
	// adhering to the specified contract and responding with HTML instead of
	// JSON-serialized data.
	// We expect that this is HTML and starts with a '<head>' and has multiple
	// lines. We want to eliminate the '\r\n' bits and present this garbage as a
	// single line, mostly for debug-ability.
	bodySlc, _ := io.ReadAll(body)
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
	}

	// At this point we may have some garbage, but let's return that anyway. :(
	return errors.New("dubious response from the API: " + string(bodySlc))
}

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
		return nil, decodeFailedApiResponse(resp)
	}

	// Handle http.StatusOK response next.
	return resp.Body, nil
}

func Authenticate(credentials *Credentials, authEndpointUrl string) (*Token, error) {
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

	req, err := http.NewRequest(
		// http.MethodPost, LyveCloudApiPrefix+"/auth/token", data)
		http.MethodPost, authEndpointUrl, data)
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
