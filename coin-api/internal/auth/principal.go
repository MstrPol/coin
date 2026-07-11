package auth

import "context"

type Role string

const (
	RoleAdmin     Role = "admin"
	RolePublisher Role = "publisher"
	RoleReader    Role = "reader"
)

type Principal struct {
	Subject    string
	Email      string
	Roles      []Role
	AuthMethod string // api_key, oidc, local
}

type ctxKey struct{}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(ctxKey{}).(Principal)
	return p, ok
}

func ParseRole(s string) (Role, bool) {
	switch Role(s) {
	case RoleAdmin, RolePublisher, RoleReader:
		return Role(s), true
	default:
		return "", false
	}
}

func (p Principal) Has(min Role) bool {
	need := roleRank(min)
	for _, r := range p.Roles {
		if roleRank(r) >= need {
			return true
		}
	}
	return false
}

func roleRank(r Role) int {
	switch r {
	case RoleAdmin:
		return 3
	case RolePublisher:
		return 2
	case RoleReader:
		return 1
	default:
		return 0
	}
}

func highestRole(roles []Role) Role {
	best := RoleReader
	bestRank := 0
	for _, r := range roles {
		if rank := roleRank(r); rank > bestRank {
			bestRank = rank
			best = r
		}
	}
	if bestRank == 0 && len(roles) == 0 {
		return RoleReader
	}
	return best
}
