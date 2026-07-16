package security

import (
	"net"
	"testing"
)

// newTestGuard builds a permissive guard (http and IP literals allowed) with a
// stubbed resolver so tests don't hit DNS. Internal-target blocking is still on.
func newTestGuard(allowPrivate bool, resolved map[string][]net.IP) *Guard {
	g := NewGuard(Config{AllowPrivate: allowPrivate, AllowHTTP: true, AllowIPLiterals: true})
	g.resolver = func(host string) ([]net.IP, error) {
		if ips, ok := resolved[host]; ok {
			return ips, nil
		}
		return nil, net.UnknownNetworkError("no such host")
	}
	return g
}

func TestValidateURL_BlocksInternalTargets(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"loopback literal", "http://127.0.0.1/"},
		{"loopback name", "http://localhost/"},
		{"ipv6 loopback", "http://[::1]/"},
		{"private 10", "http://10.0.0.5/"},
		{"private 172", "http://172.16.9.9/"},
		{"private 192", "http://192.168.1.1/"},
		{"link-local metadata", "http://169.254.169.254/latest/meta-data/"},
		{"unspecified", "http://0.0.0.0/"},
		{"cgnat", "http://100.64.1.1/"},
		{"file scheme", "file:///etc/passwd"},
		{"gopher scheme", "gopher://127.0.0.1/"},
		{"dns to private", "http://evil.example.com/"},
	}

	guard := newTestGuard(false, map[string][]net.IP{
		"localhost":        {net.ParseIP("127.0.0.1")},
		"evil.example.com": {net.ParseIP("10.1.2.3")},
	})

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := guard.ValidateURL(tc.url); err == nil {
				t.Fatalf("expected %s to be blocked, got nil error", tc.url)
			}
		})
	}
}

func TestValidateURL_AllowsPublicTargets(t *testing.T) {
	guard := newTestGuard(false, map[string][]net.IP{
		"example.com": {net.ParseIP("93.184.216.34")},
	})

	for _, url := range []string{
		"http://example.com/",
		"https://example.com/path?q=1",
		"https://93.184.216.34/",
	} {
		if err := guard.ValidateURL(url); err != nil {
			t.Fatalf("expected %s to be allowed, got %v", url, err)
		}
	}
}

func TestValidateURL_BlocksHTTPByDefault(t *testing.T) {
	g := NewGuard(Config{}) // strict defaults
	g.resolver = func(string) ([]net.IP, error) { return []net.IP{net.ParseIP("93.184.216.34")}, nil }
	if err := g.ValidateURL("http://example.com/"); err == nil {
		t.Fatal("expected http to be blocked by default")
	}
	if err := g.ValidateURL("https://example.com/"); err != nil {
		t.Fatalf("expected https to be allowed, got %v", err)
	}
}

func TestValidateURL_BlocksIPLiteralsByDefault(t *testing.T) {
	// http allowed, but IP literals still blocked by default.
	g := NewGuard(Config{AllowHTTP: true})
	g.resolver = func(string) ([]net.IP, error) { return []net.IP{net.ParseIP("93.184.216.34")}, nil }
	if err := g.ValidateURL("https://93.184.216.34/"); err == nil {
		t.Fatal("expected direct IP to be blocked by default")
	}
	if err := g.ValidateURL("http://example.com/"); err != nil {
		t.Fatalf("expected hostname to be allowed, got %v", err)
	}
}

func TestValidateURL_AllowPrivateOverride(t *testing.T) {
	guard := newTestGuard(true, map[string][]net.IP{
		"internal": {net.ParseIP("10.0.0.1")},
	})
	if err := guard.ValidateURL("http://10.0.0.1/"); err != nil {
		t.Fatalf("expected private target allowed with override, got %v", err)
	}
}

func TestValidateURL_AllowAndDenyLists(t *testing.T) {
	deny := NewGuard(Config{AllowHTTP: true, DenyHosts: []string{"blocked.example.com"}})
	deny.resolver = func(string) ([]net.IP, error) { return []net.IP{net.ParseIP("93.184.216.34")}, nil }
	if err := deny.ValidateURL("http://blocked.example.com/"); err == nil {
		t.Fatal("expected denylisted host to be blocked")
	}

	allow := NewGuard(Config{AllowHTTP: true, AllowHosts: []string{"only.example.com"}})
	allow.resolver = func(string) ([]net.IP, error) { return []net.IP{net.ParseIP("93.184.216.34")}, nil }
	if err := allow.ValidateURL("http://other.example.com/"); err == nil {
		t.Fatal("expected non-allowlisted host to be blocked")
	}
	if err := allow.ValidateURL("http://only.example.com/"); err != nil {
		t.Fatalf("expected allowlisted host to pass, got %v", err)
	}
}
