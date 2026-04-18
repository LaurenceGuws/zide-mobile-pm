package zidepm

import (
	"net"
	"strings"
)

// joinHostPortDNS53 turns a getprop-style IPv4/IPv6 literal into host:53 for DNS
// UDP/TCP. It returns "" when the value is not a usable IP (empty, interface
// garbage, etc.).
func joinHostPortDNS53(prop string) string {
	v := strings.TrimSpace(prop)
	if v == "" {
		return ""
	}
	if i := strings.IndexByte(v, '%'); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	if cut := strings.Split(v, ","); len(cut) > 0 {
		v = strings.TrimSpace(cut[0])
	}
	if ip := net.ParseIP(v); ip != nil {
		return net.JoinHostPort(ip.String(), "53")
	}
	return ""
}
