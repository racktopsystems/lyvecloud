package lyveapi

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func Test_decodeFailedApiResponse(t *testing.T) {
	t.Parallel()
	resp := &http.Response{}

	type testCase struct {
		text string
	}

	for _, testCase := range []testCase{
		{`this is not a valid JSON response`},
		{`<html><head></head><body>xxxxx\n\n\n</body>`},
	} {
		rdr := strings.NewReader(testCase.text)
		resp.Body = io.NopCloser(rdr)
		if err := decodeFailedApiResponse(resp); err == nil {
			t.Error("Expected non-nil error; got nil")
		}
	}

}
