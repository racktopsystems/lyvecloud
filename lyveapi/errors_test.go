package lyveapi

import "testing"

func TestApiCallFailedErrorMsg(t *testing.T) {

	t.Parallel()

	const msg1 = "This is a test message"
	const msg2 = "This is a test message"
	const code1 = "TestMessage"
	const code2 = "unknown"
	const code3 = "TestMessage"
	const expect1 = "request failed: " + msg1 + " " + "(" + code1 + ")"
	const expect2 = "request failed: " + msg2 + " " + "(" + code2 + ")"
	const expect3 = "request failed: no additional information given" + " " + "(" + code3 + ")"

	e1 := ApiCallFailedError{
		apiResp: &requestFailedResp{
			Code:    code1,
			Message: "This is a test message",
		},
	}

	if e1.Error() != expect1 {
		t.Errorf("actual: '%s' != expected: '%s'", e1.Error(), expect1)
	}

	e2 := ApiCallFailedError{
		apiResp: &requestFailedResp{
			Message: "This is a test message",
		},
	}

	if e2.Error() != expect2 {
		t.Errorf("actual: '%s' != expected: '%s'", e2.Error(), expect2)
	}

	e3 := ApiCallFailedError{
		apiResp: &requestFailedResp{
			Code: code3,
		},
	}
	if e3.Error() != expect3 {
		t.Errorf("actual: '%s' != expected: '%s'", e3.Error(), expect3)
	}
}
