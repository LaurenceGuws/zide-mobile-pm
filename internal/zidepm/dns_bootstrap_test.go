package zidepm

import "testing"

func TestJoinHostPortDNS53(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"  8.8.8.8  ", "8.8.8.8:53"},
		{"2001:4860:4860::8888", "[2001:4860:4860::8888]:53"},
		{"fe80::1%wlan0", "[fe80::1]:53"},
		{"192.0.2.1,192.0.2.2", "192.0.2.1:53"},
		{"not-an-ip", ""},
	}
	for _, tc := range cases {
		if got := joinHostPortDNS53(tc.in); got != tc.want {
			t.Fatalf("joinHostPortDNS53(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}
