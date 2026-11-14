package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/foyez/dbaas-platform/platform/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/http/middleware"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Middleware struct {
	mw *middleware.Interceptor[*oauth.IntrospectionContext]
}

func New(ctx context.Context, cfg config.ZitadelConfig) (*Middleware, error) {
	var opts []zitadel.Option
	if cfg.InsecurePort != "" {
		opts = append(opts, zitadel.WithInsecure(cfg.InsecurePort))
	}

	keyFile, err := client.ConfigFromKeyFile(cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("authmw: read zitadel key file: %w", err)
	}

	authZ, err := authorization.New(
		ctx,
		zitadel.New(cfg.Domain, opts...),
		// Same as oauth.DefaultAuthorization, but with a traced HTTP client
		// so each token introspection shows up as a span.
		oauth.WithIntrospection[*oauth.IntrospectionContext](
			func(ctx context.Context, issuer string) (rs.ResourceServer, error) {
				return rs.NewResourceServerJWTProfile(
					ctx, issuer, keyFile.ClientID, keyFile.KeyID, []byte(keyFile.Key),
					rs.WithClient(&http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}),
				)
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("authmw: couldn't initialize zitadel sdk: %w", err)
	}

	return &Middleware{mw: middleware.New(authZ)}, nil
}

// RequireAuth protects a handler for any authenticated user.
func (m *Middleware) RequireAuth() func(http.Handler) http.Handler {
	return m.mw.RequireAuthorization()
}

// RequireRole requires a valid access token with the specified role.
func (m *Middleware) RequireRole(role string) func(http.Handler) http.Handler {
	return m.mw.RequireAuthorization(
		authorization.WithRole(role),
	)
}

// UserID extracts the authenticated user's Zitadel subject ID from context.
// Use this inside handlers wrapped by RequireAuth/RequireRole.
func (m *Middleware) UserID(ctx context.Context) string {
	return m.mw.Context(ctx).UserID()
}

// IsGrantedRole checks whether the authenticated user has the given role.
func (m *Middleware) IsGrantedRole(ctx context.Context, role string) bool {
	return m.mw.Context(ctx).IsGrantedRole(role)
}

// Gin adapts standard net/http middleware to Gin middleware.
func Gin(mw func(http.Handler) http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		nextCalled := false

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			c.Request = r
			c.Next()
		})

		mw(next).ServeHTTP(c.Writer, c.Request)

		if !nextCalled {
			c.Abort()
		}
	}
}
