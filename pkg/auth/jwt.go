package auth

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Simple JWKS cache
type jwksKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwks struct {
	Keys []jwksKey `json:"keys"`
}

type JWKSProvider struct {
	url       string
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
	ttl       time.Duration
	issuer    string
}

func NewJWKSProvider(url, issuer string) *JWKSProvider {
	return &JWKSProvider{url: url, keys: make(map[string]*rsa.PublicKey), ttl: 10 * time.Minute, issuer: issuer}
}

func (p *JWKSProvider) getKey(kid string) (*rsa.PublicKey, error) {
	p.mu.RLock()
	if k, ok := p.keys[kid]; ok && time.Now().Before(p.expiresAt) {
		p.mu.RUnlock()
		return k, nil
	}
	p.mu.RUnlock()
	if err := p.refresh(context.Background()); err != nil {
		return nil, err
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if k, ok := p.keys[kid]; ok {
		return k, nil
	}
	return nil, fmt.Errorf("key %s not found in JWKS", kid)
}

func (p *JWKSProvider) refresh(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("jwks fetch failed: %s", resp.Status)
	}
	var j jwks
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		return err
	}
	keys := make(map[string]*rsa.PublicKey)
	for _, k := range j.Keys {
		if k.Kty != "RSA" || k.N == "" || k.E == "" || k.Kid == "" {
			continue
		}
		pub, err := parseRSAPublicKeyFromModExp(k.N, k.E)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}
	p.mu.Lock()
	p.keys = keys
	p.expiresAt = time.Now().Add(p.ttl)
	p.mu.Unlock()
	return nil
}

func parseRSAPublicKeyFromModExp(nB64URL, eB64URL string) (*rsa.PublicKey, error) {
	nBytes, err := jwt.DecodeSegment(nB64URL)
	if err != nil {
		return nil, err
	}
	eBytes, err := jwt.DecodeSegment(eB64URL)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	return &rsa.PublicKey{N: n, E: e}, nil
}

// ValidateToken parse et valide un JWT signé par Keycloak via JWKS.
func (p *JWKSProvider) ValidateToken(tokenStr string, audience string) (*jwt.Token, jwt.MapClaims, error) {
	if tokenStr == "" {
		return nil, nil, errors.New("empty token")
	}
	parser := &jwt.Parser{ValidMethods: []string{jwt.SigningMethodRS256.Name}}
	var claims jwt.MapClaims
	token, err := parser.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing kid")
		}
		return p.getKey(kid)
	})
	if err != nil || !token.Valid {
		return nil, nil, errors.New("invalid token")
	}
	// issuer
	if iss, _ := claims["iss"].(string); p.issuer != "" && iss != p.issuer {
		return nil, nil, errors.New("invalid issuer")
	}
	// aud (optional)
	if audience != "" {
		switch aud := claims["aud"].(type) {
		case string:
			if aud != audience {
				return nil, nil, errors.New("invalid audience")
			}
		case []any:
			ok := false
			for _, v := range aud {
				if s, _ := v.(string); s == audience {
					ok = true
					break
				}
			}
			if !ok {
				return nil, nil, errors.New("invalid audience")
			}
		}
	}
	return token, claims, nil
}

// BearerFromHeader extrait le token de l'en-tête Authorization.
func BearerFromHeader(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
