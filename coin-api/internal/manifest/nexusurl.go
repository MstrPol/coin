package manifest

import (
	"net/url"
	"os"
	"strings"
)

// RuntimeNexusURL rewrites localhost Nexus URLs (dev publish from host) to NEXUS_URL in stack.
func RuntimeNexusURL(stored string) string {
	if stored == "" {
		return stored
	}
	u, err := url.Parse(stored)
	if err != nil {
		return stored
	}
	host := u.Hostname()
	if host != "localhost" && host != "127.0.0.1" {
		return stored
	}
	base := os.Getenv("NEXUS_URL")
	if base == "" {
		base = "http://nexus:8081"
	}
	runtime, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return stored
	}
	u.Scheme = runtime.Scheme
	u.Host = runtime.Host
	return u.String()
}
