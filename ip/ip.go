package ip

import (
	"cmp"
	"context"
	"log/slog"
	"net/http"
	"net/netip"
	"strings"
)

// Get retrieves the external IP address from the request.
func Get(r *http.Request) string {
	return cmp.Or(r.Header.Get("X-Envoy-External-Address"), r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")])
}

// Anonymize takes an IP address as a string and returns a more anonymous version of it.
// For IPv4, it returns the first two octets
// For IPv6, it returns the first four hextets.
// If the IP is private or loopback, it returns the original IP.
// If the IP cannot be parsed, it logs a warning and returns the original IP.
func Anonymize(ctx context.Context, ip string) string {
	if len(ip) > 2 && ip[0] == '[' && ip[len(ip)-1] == ']' {
		ip = ip[1 : len(ip)-1]
	}
	parsedIP, err := netip.ParseAddr(ip)
	if err != nil {
		slog.WarnContext(ctx, "failed to parse ip", "err", err, "ip", ip)
		return ip
	}

	if parsedIP.IsPrivate() || parsedIP.IsLoopback() {
		return ip
	}

	if parsedIP.Is4() {
		prefix, err := parsedIP.Prefix(16)
		if err == nil {
			return prefix.Addr().String()
		}
		slog.WarnContext(ctx, "failed to get ip4 prefix", "err", err, "ip", ip)
	} else {
		prefix, err := parsedIP.Prefix(32)
		if err == nil {
			return prefix.Addr().String()
		}
		slog.WarnContext(ctx, "failed to get ip6 prefix", "err", err, "ip", ip)
	}

	return ip
}
