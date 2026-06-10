package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"

	"coin.local/coin-api/internal/config"
)

type OIDCVerifier struct {
	once     sync.Once
	verifier *oidc.IDTokenVerifier
	initErr  error
	cfg      config.Config
}

func NewOIDCVerifier(cfg config.Config) *OIDCVerifier {
	return &OIDCVerifier{cfg: cfg}
}

func (v *OIDCVerifier) init(ctx context.Context) error {
	v.once.Do(func() {
		if !v.cfg.OIDCEnabled || v.cfg.OIDCIssuerURL == "" {
			v.initErr = fmt.Errorf("oidc not configured")
			return
		}
		provider, err := oidc.NewProvider(ctx, v.cfg.OIDCIssuerURL)
		if err != nil {
			v.initErr = fmt.Errorf("oidc provider: %w", err)
			return
		}
		oidcCfg := &oidc.Config{SkipClientIDCheck: v.cfg.OIDCAudience == ""}
		if v.cfg.OIDCAudience != "" {
			oidcCfg.ClientID = v.cfg.OIDCAudience
		}
		v.verifier = provider.Verifier(oidcCfg)
	})
	return v.initErr
}

func (v *OIDCVerifier) Verify(ctx context.Context, rawToken string) (Principal, error) {
	if err := v.init(ctx); err != nil {
		return Principal{}, err
	}
	token, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return Principal{}, err
	}

	claimName := v.cfg.OIDCRolesClaim
	if claimName == "" {
		claimName = "roles"
	}

	var claims map[string]any
	if err := token.Claims(&claims); err != nil {
		return Principal{}, err
	}

	roles := extractRoles(claims[claimName])
	if len(roles) == 0 {
		roles = []Role{RoleReader}
	}

	email, _ := claims["email"].(string)
	subject := token.Subject
	if subject == "" {
		subject, _ = claims["sub"].(string)
	}

	return Principal{
		Subject:    subject,
		Email:      email,
		Roles:      roles,
		AuthMethod: "oidc",
	}, nil
}

func extractRoles(raw any) []Role {
	switch v := raw.(type) {
	case []any:
		var roles []Role
		for _, item := range v {
			if s, ok := item.(string); ok {
				if r, ok := ParseRole(s); ok {
					roles = append(roles, r)
				}
			}
		}
		return roles
	case []string:
		var roles []Role
		for _, s := range v {
			if r, ok := ParseRole(s); ok {
				roles = append(roles, r)
			}
		}
		return roles
	default:
		return nil
	}
}

func looksLikeJWT(token string) bool {
	return strings.Count(token, ".") == 2
}
