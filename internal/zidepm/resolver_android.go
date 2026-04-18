package zidepm

import (
	"context"
	"net"
	"os/exec"
	"strings"
	"time"
)

func init() {
	// Android app processes often ship a resolv.conf that points at loopback DNS
	// (e.g. [::1]:53) which is unreachable from the embedded zide-pm binary built
	// with CGO_ENABLED=0. Prefer explicit system properties, then public DNS, so
	// GitHub manifest and artifact fetches work without root or manual resolver
	// configuration (APX-B18 / MP-A10).
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial:     androidDNSDial,
	}
}

func androidDNSDial(ctx context.Context, network string, _ string) (net.Conn, error) {
	d := net.Dialer{Timeout: 8 * time.Second}
	hostport := androidNameserverHostPort(ctx)
	nw := "udp"
	if strings.HasPrefix(network, "tcp") {
		nw = "tcp"
	}
	return d.DialContext(ctx, nw, hostport)
}

func androidNameserverHostPort(ctx context.Context) string {
	for _, key := range []string{"net.dns1", "net.dns2", "net.dns3", "net.dns4"} {
		if v := androidGetprop(ctx, key); v != "" {
			if hp := joinHostPortDNS53(v); hp != "" {
				return hp
			}
		}
	}
	return "8.8.8.8:53"
}

func androidGetprop(ctx context.Context, key string) string {
	propCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()

	exePaths := []string{"/system/bin/getprop", "/system/xbin/getprop"}
	for _, exe := range exePaths {
		out, err := exec.CommandContext(propCtx, exe, key).Output()
		if err != nil {
			continue
		}
		return strings.TrimSpace(string(out))
	}
	if path, err := exec.LookPath("getprop"); err == nil {
		out, err := exec.CommandContext(propCtx, path, key).Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return ""
}
