package dbaas

import (
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// WithClientCredentials configures the SDK to authenticate against ZITADEL
// using the OAuth2 client credentials grant, for machine users (CI, scripts,
// other services). Tokens are fetched lazily on the first request, cached,
// and fetched again shortly before they expire.
//
// scopes must include the project audience scope
// "urn:zitadel:iam:org:project:id:<projectID>:aud", or the API can't
// validate the token. Add "urn:zitadel:iam:org:project:roles" to put the
// user's roles in the token.
func WithClientCredentials(tokenURL, clientID, clientSecret string, scopes []string) Option {
	return func(c *Client) {
		cfg := &clientcredentials.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			TokenURL:     tokenURL,
			Scopes:       scopes,
		}

		// cfg.TokenSource already wraps itself in a ReuseTokenSource, so
		// tokens are cached until they're about to expire.
		c.newTokenSource = func(hc *http.Client) oauth2.TokenSource {
			return cfg.TokenSource(tokenContext(hc))
		}
	}
}

// WithToken authenticates every request with a fixed bearer token, e.g. one
// a user obtained by logging in to the web UI. The SDK can't refresh it:
// once it expires, requests fail with ErrUnauthorized.
func WithToken(token string) Option {
	return func(c *Client) {
		c.newTokenSource = func(*http.Client) oauth2.TokenSource {
			return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		}
	}
}
