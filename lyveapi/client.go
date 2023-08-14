package lyveapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type tokenDetails struct {
	// Contains the secret which was exchanged for valid account credentials.
	token string
	// After this time the token will expire and must be renewed.
	expiresAfter time.Time
	// Timestamp of token issuance by the API.
	issuedTimestamp time.Time
}

// Client is the structure used for interaction with the Lyve Cloud API through
// its methods. This structure and its public methods are expected to be
// thread-safe.
type Client struct {
	apiUrl string // url used as the entrypoint into the Lyve Cloud API
	mtx    sync.RWMutex
	tokenDetails
}

// NewCredentials returns a pointer to an initialized Credentials structure.
// This structure is an expected input into the NewClient function which returns
// a fully initialized client.
func NewCredentials(accountId, accessKey, secret string) *Credentials {
	return &Credentials{
		AccountId: accountId,
		AccessKey: accessKey,
		Secret:    secret,
	}
}

// NewClient initializes a Lyve Cloud API client and performs an authentication,
// which if successful results in the API returning a token, with which the
// client is initialized. Without this token any further interactions with the
// API will not be possible. On success, this function returns a pointer to
// an initialized Client struct and nil error. On failure, the function returns
// a nil instead of a pointer to Client and an error.
//
// Arguments:
//
// credentials are exchanged with the API for an API token.
// apiUrl is the base endpoint URL for the API. If an empty string is supplied,
// we fallback to the default base endpoint URL. This parameter is primarily
// useful for testing and not everyday production operations.
func NewClient(credentials *Credentials, apiUrl string) (*Client, error) {
	const roundTo = nsecPerSec

	// If there is nothing passed-in for apiUrl, use default LyveCloud API URL.
	if apiUrl == "" {
		apiUrl = LyveCloudApiPrefix
	}

	var auth *Token
	var authEndpointUrl = apiUrl + "/auth/token"
	var err error
	var tokValidForSeconds int

	if auth, err = Authenticate(credentials, authEndpointUrl); err != nil {
		return nil, err
	}

	now := time.Now()

	if tokValidForSeconds, err = strconv.Atoi(auth.ExpirationSec); err != nil {
		return nil, err
	}

	// This is imprecise for a few reasons. First, we are rounding here, and
	// second receiving token validity from the API in seconds, which we then
	// add to a timestamp that we took, which is hopefully close, but not
	// guaranteed to be close to the clocks on the API side. Finally, the
	// validity period returned by the API is low-precision, seconds as opposed
	// to msecs, usecs, etc.
	tokExpiresAfter := now.Add(
		time.Duration(tokValidForSeconds * nsecPerSec)).Round(roundTo)

	return &Client{
		apiUrl,
		sync.RWMutex{},
		tokenDetails{
			token:           auth.Token,
			expiresAfter:    tokExpiresAfter,
			issuedTimestamp: now,
		},
	}, nil
}

// Token is a string representation of the token previously returned by thr API
// in exchange for valid account credentials.
func (client *Client) Token() string {
	client.mtx.RLock()
	token := client.token
	client.mtx.RUnlock()
	return token
}

// ExpiresAfter is an estimate of when the token will no-longer be valid and
// require renewal or re-issuance.
func (client *Client) ExpiresAfter() time.Time {
	client.mtx.RLock()
	expiresAfter := client.expiresAfter
	client.mtx.RUnlock()
	return expiresAfter
}

// TokenExpired returns true when the token is believed to be expired.
func (client *Client) TokenExpired() bool {
	client.mtx.RLock()
	expired := time.Now().After(client.expiresAfter)
	client.mtx.RUnlock()
	return expired
}

// TokenValidUntil returns the end of this token's validity as a time.Time
// value. Due to the nature of clocks, issues such as drift, corrections, etc.
// will result in some inherent inaccuracy of this information, since the
// returned object is based on the state of the system's clock at the moment
// the API is queried and may subsequently be adjusted due to drift, etc.
func (client *Client) TokenValidUntil() (time.Time, error) {
	var err error
	var expiresInSecs int
	var rdr io.ReadCloser

	client.mtx.RLock()
	endpoint := client.apiUrl + "/auth/token"
	token := client.token
	client.mtx.RUnlock()

	now := time.Now()

	if rdr, err = apiRequestAuthenticated(
		token, http.MethodGet, endpoint, nil); err != nil {
		return time.Time{}, err
	}

	defer rdr.Close()

	tokValidResp := &Token{}
	respBodyDecoder := json.NewDecoder(rdr)
	if err = respBodyDecoder.Decode(tokValidResp); err != nil {
		return time.Time{}, err
	}

	if expiresInSecs, err =
		strconv.Atoi(tokValidResp.ExpirationSec); err != nil {
		return time.Time{}, err
	}

	return now.Add(
		time.Duration(expiresInSecs * nsecPerSec)).Round(nsecPerSec), nil
}

// SetApiURL is really not intended for production use but exists to ease
// certain testing aspects. If you are using it outside of testing, you should
// think again about the implementation.
func (client *Client) SetApiURL(apiUrl string) {
	client.mtx.Lock()
	client.apiUrl = apiUrl
	client.mtx.Unlock()
}
