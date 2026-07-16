// Package security provides request-time protections, most importantly an
// SSRF guard that prevents the impersonation proxy from being pointed at
// internal or cloud-metadata addresses.
package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// Guard validates outbound request targets to block SSRF.
type Guard struct {
	// AllowPrivate disables the private/loopback/link-local IP checks. Intended
	// for deployments that intentionally proxy to internal hosts.
	AllowPrivate bool
	// DenyHosts is a list of exact hostnames that are always rejected.
	DenyHosts map[string]struct{}
	// AllowHosts, when non-empty, restricts requests to these exact hostnames
	// only (allowlist mode).
	AllowHosts map[string]struct{}
	// resolver is pluggable for testing.
	resolver func(host string) ([]net.IP, error)
}

// NewGuard builds a Guard from config values.
func NewGuard(allowPrivate bool, denyHosts, allowHosts []string) *Guard {
	toSet := func(items []string) map[string]struct{} {
		m := make(map[string]struct{}, len(items))
		for _, it := range items {
			it = strings.TrimSpace(strings.ToLower(it))
			if it != "" {
				m[it] = struct{}{}
			}
		}
		return m
	}
	return &Guard{
		AllowPrivate: allowPrivate,
		DenyHosts:    toSet(denyHosts),
		AllowHosts:   toSet(allowHosts),
		resolver: func(host string) ([]net.IP, error) {
			return net.LookupIP(host)
		},
	}
}

// ValidateURL parses and validates a target URL, returning an error describing
// why the destination is not allowed. The error message is safe to return to
// clients.
func (g *Guard) ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("scheme not allowed: only http and https are permitted")
	}

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return fmt.Errorf("invalid URL: missing host")
	}

	if len(g.AllowHosts) > 0 {
		if _, ok := g.AllowHosts[host]; !ok {
			return fmt.Errorf("host not in allowlist: %s", host)
		}
	}
	if _, denied := g.DenyHosts[host]; denied {
		return fmt.Errorf("host is denied: %s", host)
	}

	// If the host is a literal IP, validate it directly.
	if ip := net.ParseIP(host); ip != nil {
		return g.checkIP(ip, host)
	}

	if g.AllowPrivate {
		return nil
	}

	// Resolve and validate every address the host maps to, to defend against
	// DNS records that point at internal ranges.
	ips, err := g.resolver(host)
	if err != nil {
		return fmt.Errorf("could not resolve host: %s", host)
	}
	if len(ips) == 0 {
		return fmt.Errorf("could not resolve host: %s", host)
	}
	for _, ip := range ips {
		if err := g.checkIP(ip, host); err != nil {
			return err
		}
	}
	return nil
}

// checkIP rejects addresses that target the host itself or internal networks.
func (g *Guard) checkIP(ip net.IP, host string) error {
	if g.AllowPrivate {
		return nil
	}
	if isBlockedIP(ip) {
		return fmt.Errorf("destination address is not allowed: %s", host)
	}
	return nil
}

// isBlockedIP reports whether an IP belongs to a range that must never be
// reachable through the proxy.
func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsInterfaceLocalMulticast() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		// RFC1918 private ranges.
		if ip4[0] == 10 {
			return true
		}
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		// 100.64.0.0/10 carrier-grade NAT.
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return true
		}
		// 192.0.0.0/24 IETF protocol assignments (includes some metadata proxies).
		if ip4[0] == 192 && ip4[1] == 0 && ip4[2] == 0 {
			return true
		}
		return false
	}
	// IPv6 unique local addresses fc00::/7.
	if len(ip) == net.IPv6len && (ip[0]&0xfe) == 0xfc {
		return true
	}
	return false
}
