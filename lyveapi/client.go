package lyveapi

import (
	"context"
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
	issuedTimestamp  time.Time
	expiresMonoNanos time.Duration
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
// This function does not allow the consumer to pass a context to the
// underlying HTTP client. If the consumer wants to pass a context to the http
// client, please use NewClientWithContext factory function instead of this
// function.
//
// Arguments:
//
// credentials -- API authentication creds which the API will exchange for an
// API token with a limited lifetime.
//
// apiUrl -- Base endpoint URL for the Lyve Cloud API. If an empty string is
// supplied, we fallback to the default Lyve Cloud API base endpoint URL. This
// parameter is primarily useful for testing and not everyday production
// operations.
func NewClient(credentials *Credentials, apiUrl string) (*Client, error) {
	return newClientImpl(context.Background(), credentials, apiUrl)
}

// NewClientWithContext is a factory function which is functionally identical
// to NewClient(...) with the only difference being the context parameter as
// the first argument. See documentation for the NewClient(...) factory
// function for usage details.
func NewClientWithContext(
	ctx context.Context, credentials *Credentials, apiUrl string) (*Client, error) {
	return newClientImpl(ctx, credentials, apiUrl)
}

func newClientImpl(
	ctx context.Context, credentials *Credentials, apiUrl string) (*Client, error) {
	const roundTo = nsecPerSec

	// If there is nothing passed-in for apiUrl, use default Lyve Cloud API URL.
	if apiUrl == "" {
		apiUrl = LyveCloudApiPrefix
	}

	var auth *Token
	var authEndpointUrl = apiUrl
	var err error
	var tokValidForSeconds int

	if auth, err = Authenticate(ctx, credentials, authEndpointUrl); err != nil {
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

// NewAuthenticatedClient initializes a Lyve Cloud API client without first
// authenticating with the API. The supplied token is assumed to be valid,
// however the API is queried to determine the expiry of the token. If that
// query fails, which would typically happen if the token was already expired
// that error will be returned to the caller with a nil instead of a pointer to
// an initialized *Client. Otherwise we will return an initialized client and a
// nil error. This client will be usable for at least the
func NewAuthenticatedClient(token, apiUrl string) (*Client, error) {
	const roundTo = nsecPerSec

	var err error
	var expiresIn time.Duration

	// If there is nothing passed-in for apiUrl, use default LyveCloud API URL.
	if apiUrl == "" {
		apiUrl = LyveCloudApiPrefix
	}

	now := time.Now()

	if expiresIn, err = getTokenExpiresDuration(token, apiUrl); err != nil {
		return nil, err
	}

	// This is imprecise for a few reasons. First, we are rounding here, and
	// second receiving token validity from the API in seconds, which we then
	// add to a timestamp that we took, which is hopefully close, but not
	// guaranteed to be close to the clocks on the API side. Finally, the
	// validity period returned by the API is low-precision, seconds as opposed
	// to msecs, usecs, etc.
	tokExpiresAfter := now.Add(expiresIn).Round(roundTo)

	return &Client{
		apiUrl,
		sync.RWMutex{},
		tokenDetails{
			token:           token,
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

// func (client *Client) ExpiresAfterMonoNanos() time.Duration {
// 	client.mtx.RLock()
// 	expiresAfter := client.expiresMonoNanos
// 	client.mtx.RUnlock()
// 	return expiresAfter
// }

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
	var expiresIn time.Duration

	client.mtx.RLock()
	apiUrl := client.apiUrl
	token := client.token
	client.mtx.RUnlock()

	now := time.Now()

	if expiresIn, err = getTokenExpiresDuration(token, apiUrl); err != nil {
		return time.Time{}, err
	}

	return now.Add(expiresIn).Round(nsecPerSec), nil
}

func getTokenExpiresDuration(token, apiUrl string) (time.Duration, error) {
	var err error
	var expiresInSecs int
	var rdr io.ReadCloser

	var endpoint string
	if apiUrl != "" {
		endpoint = apiUrl + "/auth/token"
	} else {
		endpoint = LyveCloudApiPrefix + "/auth/token"
	}

	if rdr, err = apiRequestAuthenticated(
		token, http.MethodGet, endpoint, nil); err != nil {
		return 0, err
	}

	defer rdr.Close()

	tok := Token{}
	respBodyDecoder := json.NewDecoder(rdr)
	if err = respBodyDecoder.Decode(&tok); err != nil {
		return 0, err
	}

	if expiresInSecs, err =
		strconv.Atoi(tok.ExpirationSec); err != nil {
		return 0, err
	}

	return time.Second * time.Duration(expiresInSecs), nil
}

// // getTokenExpiresSeconds returns the number of seconds relative to some
// // definition of "now", according to the server, during which the token remains
// // valid.
// func getTokenExpiresSeconds(token, apiUrl string) (time.Duration, error) {
// 	var err error
// 	var rdr io.ReadCloser

// 	var endpoint string
// 	if apiUrl != "" {
// 		endpoint = apiUrl + "/auth/token"
// 	} else {
// 		endpoint = LyveCloudApiPrefix + "/auth/token"
// 	}

// 	if rdr, err = apiRequestAuthenticated(
// 		token, http.MethodGet, endpoint, nil); err != nil {
// 		return 0, err
// 	}

// 	defer rdr.Close()

// 	tok := Token{}
// 	respBodyDecoder := json.NewDecoder(rdr)
// 	if err = respBodyDecoder.Decode(&tok); err != nil {
// 		return 0, err
// 	}

// 	return tok.ExpiresMonoNanos()
// }

// SetApiURL is really not intended for production use but exists to ease
// certain testing aspects. If you are using it outside of testing, you should
// think again about the implementation.
func (client *Client) SetApiURL(apiUrl string) {
	client.mtx.Lock()
	client.apiUrl = apiUrl
	client.mtx.Unlock()
}
