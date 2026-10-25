// Package clientip resolves client addresses using an explicit proxy trust policy.
package clientip

import (
	"errors"
	"net/http"
	"net/netip"
	"strings"
)

const (
	maxForwardedBytes = 4096
	maxForwardedHops  = 32
)

// Policy is immutable after parsing. Its zero value trusts no proxy.
type Policy struct {
	prefixes []netip.Prefix
}

// Parse accepts comma-separated IP literals or CIDRs. Errors are safe to log:
// they identify the setting without including its contents.
func Parse(raw string) (Policy, error) {
	var policy Policy
	if strings.TrimSpace(raw) == "" {
		return policy, nil
	}
	invalid := errors.New("TRUSTED_PROXIES must contain valid IP addresses or CIDR ranges without zones; mapped IPv4 prefixes must be at least /96")
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		var prefix netip.Prefix
		if strings.Contains(entry, "/") {
			var err error
			prefix, err = netip.ParsePrefix(entry)
			if err != nil {
				return Policy{}, invalid
			}
		} else {
			addr, err := netip.ParseAddr(entry)
			if err != nil || addr.Zone() != "" {
				return Policy{}, invalid
			}
			prefix = netip.PrefixFrom(addr, addr.BitLen())
		}
		if prefix.Addr().Is4In6() {
			if prefix.Bits() < 96 {
				return Policy{}, invalid
			}
			prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
		}
		policy.prefixes = append(policy.prefixes, prefix.Masked())
	}
	return policy, nil
}

// Resolve returns a canonical client IP without modifying the request. Only
// trusted peers may supply X-Forwarded-For. The first untrusted hop from the
// right ends the search, so an attacker-controlled prefix cannot choose the IP.
// Malformed or oversized trusted chains fall back to the connection peer.
// An invalid peer produces the zero address, a shared fallback key for callers.
func (p Policy) Resolve(r *http.Request) netip.Addr {
	var peer netip.Addr
	if endpoint, err := netip.ParseAddrPort(r.RemoteAddr); err == nil {
		peer = endpoint.Addr()
	} else {
		peer, _ = netip.ParseAddr(r.RemoteAddr)
	}
	peer = peer.WithZone("").Unmap()
	if !p.trusts(peer) {
		return peer
	}
	values := r.Header.Values("X-Forwarded-For")
	if len(values) == 0 {
		return peer
	}
	// Count the commas implied between repeated fields as well as their values.
	total := len(values) - 1
	if total > maxForwardedBytes {
		return peer
	}
	for _, value := range values {
		if len(value) > maxForwardedBytes-total {
			return peer
		}
		total += len(value)
	}
	last, hops := peer, 0
	for i := len(values) - 1; i >= 0; i-- {
		rest := values[i]
		for {
			if hops == maxForwardedHops {
				return peer
			}
			hops++
			comma := strings.LastIndexByte(rest, ',')
			addr, err := netip.ParseAddr(strings.Trim(rest[comma+1:], " \t"))
			if err != nil || addr.Zone() != "" {
				return peer
			}
			last = addr.Unmap()
			if !p.trusts(last) {
				return last
			}
			if comma < 0 {
				break
			}
			rest = rest[:comma]
		}
	}
	return last
}

func (p Policy) trusts(addr netip.Addr) bool {
	for _, prefix := range p.prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}
