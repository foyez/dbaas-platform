package dbaas

import (
	"context"

	"golang.org/x/oauth2/clientcredentials"
)

// WithClientCredentials configures the SDK to authenticate against ZITADEL
// using the OAuth2 client credentials grant. Tokens are fetched lazily on
// first request and cached/refreshed automatically
func WithClientCredentials(tokenURL, clientID, clientSecret string, scopes []string) Option {
	return func(c *Client) {
		cfg := &clientcredentials.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			TokenURL:     tokenURL,
			Scopes:       scopes,
		}

		// oauth2 wraps our http.Client so every outgoing request
		// automatically gets a valid Authorization header, refreshed
		// behind the scenes when the cached token expires.
		c.httpClient = cfg.Client(context.Background())
	}
}
