package clientip_test

import (
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"gotest/internal/clientip"
)

func TestResolveFollowsOnlyTheTrustedSuffix(t *testing.T) {
	for _, tc := range []struct {
		name, trusted, peer string
		forwarded           []string
		want                string
	}{
		{"default ignores spoofing", "", "192.0.2.5:1000", []string{"198.51.100.1"}, "192.0.2.5"},
		{"loopback is not implicitly trusted", "", "127.0.0.1:1000", []string{"198.51.100.1"}, "127.0.0.1"},
		{"private peer is not implicitly trusted", "", "10.0.0.2:1000", []string{"198.51.100.1"}, "10.0.0.2"},
		{"untrusted peer ignores spoofing", "10.0.0.0/8", "192.0.2.5:1000", []string{"198.51.100.1, 10.0.0.1"}, "192.0.2.5"},
		{"single proxy", "10.0.0.2", "10.0.0.2:1000", []string{"198.51.100.1"}, "198.51.100.1"},
		{"multiple proxies", "10.0.0.0/8, 2001:db8:1::/48", "10.0.0.2:1000", []string{"198.51.100.1, 2001:db8:1::5, 10.1.1.1"}, "198.51.100.1"},
		{"stop at untrusted proxy", "10.0.0.0/8", "10.0.0.2:1000", []string{"198.51.100.1, 192.0.2.5, 10.1.1.1"}, "192.0.2.5"},
		{"ignore malformed prefix", "10.0.0.0/8", "10.0.0.2:1000", []string{"unknown,,garbage, 198.51.100.1, 10.1.1.1"}, "198.51.100.1"},
		{"repeated header fields", "10.0.0.0/8", "10.0.0.2:1000", []string{"198.51.100.1", "192.0.2.5, 10.1.1.1"}, "192.0.2.5"},
		{"trusted suffix across header fields", "10.0.0.0/8", "10.0.0.2:1000", []string{"198.51.100.1, 10.2.2.2", "10.1.1.1"}, "198.51.100.1"},
		{"empty untrusted prefix is ignored", "10.0.0.0/8", "10.0.0.2:1000", []string{"", "198.51.100.1"}, "198.51.100.1"},
		{"all hops trusted", "10.0.0.0/8", "10.0.0.2:1000", []string{"10.2.2.2, 10.1.1.1"}, "10.2.2.2"},
		{"whitespace and host bits in config", " 10.0.0.5/24 , 2001:db8:1::5/48 ", "10.0.0.2:1000", []string{" \t198.51.100.1\t , 2001:db8:1::8 "}, "198.51.100.1"},
		{"IPv6 client", "10.0.0.0/8", "10.0.0.2:1000", []string{"2001:0db8:0:0:0:0:0:5"}, "2001:db8::5"},
		{"IPv6 proxy", "2001:db8:1::/48", "[2001:db8:1::2]:1000", []string{"198.51.100.1"}, "198.51.100.1"},
		{"mapped IPv4 peer", "10.0.0.0/8", "[::ffff:10.0.0.2]:1000", []string{"198.51.100.1"}, "198.51.100.1"},
		{"mapped IPv4 client", "10.0.0.0/8", "10.0.0.2:1000", []string{"::ffff:198.51.100.1"}, "198.51.100.1"},
		{"mapped IPv4 proxy in chain", "10.0.0.0/8", "10.0.0.2:1000", []string{"198.51.100.1, ::ffff:10.1.1.1"}, "198.51.100.1"},
		{"mapped IPv4 network", "::ffff:10.0.0.0/104", "10.0.0.2:1000", []string{"198.51.100.1"}, "198.51.100.1"},
		{"mapped IPv4 literal", "::ffff:10.0.0.2", "10.0.0.2:1000", []string{"198.51.100.1"}, "198.51.100.1"},
		{"IPv6 range does not trust IPv4", "::/0", "[::ffff:192.0.2.5]:1000", []string{"198.51.100.1"}, "192.0.2.5"},
		{"zoned TCP peer", "fe80::/10", "[fe80::2%eth0]:1000", []string{"198.51.100.1"}, "198.51.100.1"},
		{"bare peer", "10.0.0.2", "10.0.0.2", []string{"198.51.100.1"}, "198.51.100.1"},
		{"bare IPv6 peer", "", "2001:0db8:0:0:0:0:0:5", nil, "2001:db8::5"},
		{"invalid peer ignores headers", "0.0.0.0/0, ::/0", "not-an-address", []string{"198.51.100.1"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policy, err := clientip.Parse(tc.trusted)
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tc.peer
			for _, value := range tc.forwarded {
				r.Header.Add("X-Forwarded-For", value)
			}
			var want netip.Addr
			if tc.want != "" {
				want = netip.MustParseAddr(tc.want)
			}
			if got := policy.Resolve(r); got != want {
				t.Fatalf("client IP = %v; want %v", got, want)
			}
			if r.RemoteAddr != tc.peer {
				t.Fatal("resolving client IP changed the transport peer")
			}
		})
	}
}

func TestResolveFallsBackOnUnusableForwardedChain(t *testing.T) {
	policy, err := clientip.Parse("10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		values []string
	}{
		{"missing", nil}, {"empty", []string{""}}, {"whitespace", []string{" \t "}},
		{"unknown", []string{"198.51.100.1, unknown, 10.1.1.1"}},
		{"empty trusted hop", []string{"198.51.100.1,,10.1.1.1"}},
		{"empty last field", []string{"198.51.100.1", ""}},
		{"IPv4 with port", []string{"198.51.100.1:1000"}},
		{"IPv6 with port", []string{"[2001:db8::1]:1000"}},
		{"bracketed IPv6", []string{"[2001:db8::1]"}},
		{"zone identifier", []string{"2001:db8::1%eth0"}},
		{"non-HTTP whitespace", []string{"\u00a0198.51.100.1"}},
		{"Forwarded syntax", []string{"for=198.51.100.1"}},
		{"oversized field", []string{strings.Repeat("x", 4097) + ", 198.51.100.1"}},
		{"oversized combined fields", []string{strings.Repeat("x", 2100), strings.Repeat("x", 2100) + ", 198.51.100.1"}},
		{"too many inspected hops", []string{"198.51.100.1" + strings.Repeat(", 10.1.1.1", 32)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = "10.0.0.2:1000"
			for _, value := range tc.values {
				r.Header.Add("X-Forwarded-For", value)
			}
			r.Header.Set("X-Real-IP", "198.51.100.2")
			r.Header.Set("Forwarded", "for=198.51.100.3")
			if got := policy.Resolve(r); got != netip.MustParseAddr("10.0.0.2") {
				t.Fatalf("unusable chain resolved to %v instead of peer", got)
			}
		})
	}
}

func TestResolveAcceptsBoundariesAndIgnoresUntrustedHops(t *testing.T) {
	policy, err := clientip.Parse("10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}
	client := "198.51.100.1"
	for _, value := range []string{
		client + strings.Repeat(",10.1.1.1", 31),
		strings.Repeat("x,", 40) + client,
		strings.Repeat("x", 4096-len(client)-1) + "," + client,
	} {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "10.0.0.2:1000"
		r.Header.Set("X-Forwarded-For", value)
		if got := policy.Resolve(r); got != netip.MustParseAddr(client) {
			t.Fatalf("valid chain within bounds resolved to %v", got)
		}
	}
}

func TestParseRejectsInvalidTrustWithoutDisclosingValues(t *testing.T) {
	for _, raw := range []string{
		"private-proxy.example.test", "http://10.0.0.1", "10.0.0.1:8080", "[::1]", "fe80::1%eth0",
		"10.0.0.0/33", "2001:db8::/129", "::ffff:10.0.0.0/80", "fe80::1%eth0/64",
		",", "10.0.0.1,", ",10.0.0.1", "10.0.0.1, ,10.0.0.2", "10.0.0.1 10.0.0.2",
		"10.0.0.1,private-proxy.example.test",
	} {
		t.Run(raw, func(t *testing.T) {
			policy, err := clientip.Parse(raw)
			if err == nil {
				t.Fatal("invalid trust configuration accepted")
			}
			if !strings.Contains(err.Error(), "TRUSTED_PROXIES") || strings.Contains(err.Error(), raw) {
				t.Fatalf("expected safe configuration error; got %q", err)
			}
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = "10.0.0.1:1000"
			r.Header.Set("X-Forwarded-For", "198.51.100.1")
			if got := policy.Resolve(r); got != netip.MustParseAddr("10.0.0.1") {
				t.Fatal("failed parsing retained partially trusted ranges")
			}
		})
	}
}

func TestZeroPolicyAndEmptyConfigurationTrustNoProxy(t *testing.T) {
	policies := []clientip.Policy{{}}
	for _, raw := range []string{"", " \t\n "} {
		policy, err := clientip.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		policies = append(policies, policy)
	}
	for _, policy := range policies {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "[::1]:1000"
		r.Header.Set("X-Forwarded-For", "198.51.100.1")
		if got := policy.Resolve(r); got != netip.MustParseAddr("::1") {
			t.Fatalf("empty policy trusted a proxy: %v", got)
		}
	}
}
