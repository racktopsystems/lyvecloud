package lyveapi

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type badReader struct{}

func (r *badReader) Read(_ []byte) (int, error) {
	return 0, errors.New("boom")
}

func Test_decodeFailedApiResponse(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		errMsg string
		resp   *http.Response
	}

	for _, testCase := range []testCase{
		{
			name:   "text-instead-of-JSON",
			errMsg: `dubious response from the API: unable to decode response body`,
			resp: &http.Response{
				Body: io.NopCloser(strings.NewReader(`this is not a valid JSON response`)),
			},
		},
		{
			name:   "HTML-instead-of-JSON",
			errMsg: `dubious response from the API: <html><head></head><body>xxxxx</body>`,
			resp: &http.Response{
				Body: io.NopCloser(strings.NewReader(`<html><head></head><body>xxxxx\n\n\n</body>`)),
			},
		},
		{
			name:   "valid-API-error-response",
			errMsg: `request failed: This is a mock error message (InvalidArgument)`,
			resp: &http.Response{
				Body: io.NopCloser(strings.NewReader(`{"code": "InvalidArgument", "message": "This is a mock error message"}`)),
			},
		},
		{
			name:   "nil-body-reader",
			errMsg: `dubious response from the API: unable to decode response body`,
			resp: &http.Response{
				Body: io.NopCloser(bytes.NewReader(nil)),
			},
		},
		{
			name:   "always-failing-reader",
			errMsg: `dubious response from the API: unable to decode response body`,
			resp: &http.Response{
				Body: io.NopCloser(&badReader{}),
			},
		},
	} {
		t.Run(testCase.name, func(tt *testing.T) {
			if err := decodeFailedApiResponse(testCase.resp); err.Error() != testCase.errMsg {
				t.Errorf("Expected error: %v; actual error: %v", testCase.errMsg, err)
			}
		})
	}
}
