package middleware

import (
	"net/http"

	"github.com/flotio-dev/project-service/pkg/auth"
	"github.com/flotio-dev/project-service/pkg/httpx"
)

// RequireAuth vérifie le token Bearer via Keycloak JWKS et injecte sub/claims dans le contexte.
func RequireAuth(p *auth.JWKSProvider, audience string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := auth.BearerFromHeader(r)
			if tokenStr == "" {
				httpx.Unauthorized(w, "missing bearer token")
				return
			}
			_, claims, err := p.ValidateToken(tokenStr, audience)
			if err != nil {
				httpx.Unauthorized(w, err.Error())
				return
			}
			sub, _ := claims["sub"].(string)
			r = WithValue(r, ctxKeyToken, tokenStr)
			r = WithValue(r, ctxKeyClaims, claims)
			r = WithValue(r, ctxKeySub, sub)
			next.ServeHTTP(w, r)
		})
	}
}
