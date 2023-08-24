# Seagate Lyve Cloud Go API client
This repository is an early _work in progress_ Seagate Lyve Cloud API client written in Go.

## Getting Started
Everything starts with initializing a client. There are two supported starting points. If an authentication token has been obtained previously and is known it could be used with the `lyveapi.NewAuthenticatedClient(...)` function. Initializing the client without a previously obtained token requires population of a `lyveapi.Credentials` struct with the account name, access key and secret key and then passing this information to `lyveapi.NewClient(...)`. The second argument to this function is a base URI, which under normal circumstances should be unnecessary. Thus, simply pass in an empty string`""`.

This snippet is an example for how a new client is initialized:
```
	cred := lyveapi.NewCredentials(
		"dummyaccount",
		"dummy-access-key",
		"dummy-secret-key",
	)

	client, err := lyveapi.NewClient(cred, "")
	if err != nil {
		// Do something with the error
		return err
	}
```

Once you have a client, all further operations are methods on the client. Methods are lightly documented, but they can use better documentation to be sure.