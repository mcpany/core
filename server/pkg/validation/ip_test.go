package validation

import (
	"net"
	"testing"
)

func TestIsNAT64(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"64:ff9b::192.0.2.1", true}, // Well-known prefix
		{"64:ff9b::127.0.0.1", true},
		{"64:ff9b::1", true},         // Technically valid in prefix
		{"::1", false},
		{"2001:db8::1", false},
		{"192.0.2.1", false},
		{"64:ff9b::1:2:3", false},    // 1:2:3 implies bits in 80-96 range are non-zero
		{"65:ff9b::1", false},
		{"64:ff9c::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if got := IsNAT64(ip); got != tt.want {
				t.Errorf("IsNAT64(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsIPv4Compatible(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"::192.0.2.1", true}, // IPv4 compatible
		{"::7f00:0001", true}, // ::127.0.0.1
		{"::1", true},         // ::1 matches the pattern (96 bits of zeros), represents 0.0.0.1
		{"::0.0.0.1", true},
		{"::ffff:192.0.2.1", false}, // IPv4-mapped, not compatible
		{"2001:db8::1", false},
		{"192.0.2.1", false}, // IPv4 is not IPv6
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if got := IsIPv4Compatible(ip); got != tt.want {
				t.Errorf("IsIPv4Compatible(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsNAT64LinkLocal(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"64:ff9b::169.254.1.1", true},
		{"64:ff9b::a9fe:0101", true},   // hex for 169.254.1.1
		{"64:ff9b::169.255.1.1", false},
		{"64:ff9b::127.0.0.1", false},
		{"2001:db8::169.254.1.1", false},
		{"169.254.1.1", false}, // Not NAT64
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if got := IsNAT64LinkLocal(ip); got != tt.want {
				t.Errorf("IsNAT64LinkLocal(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsNAT64Loopback(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"64:ff9b::127.0.0.1", true},
		{"64:ff9b::7f00:0001", true}, // hex for 127.0.0.1
		{"64:ff9b::127.1.1.1", true}, // Entire 127/8 block
		{"64:ff9b::128.0.0.1", false},
		{"64:ff9b::169.254.1.1", false},
		{"2001:db8::127.0.0.1", false},
		{"127.0.0.1", false}, // Not NAT64
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if got := IsNAT64Loopback(ip); got != tt.want {
				t.Errorf("IsNAT64Loopback(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsPrivateNetworkIPv4(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		// 0.0.0.0/8
		{"0.0.0.0", true},
		{"0.1.1.1", true},
		{"1.0.0.0", false},

		// 10.0.0.0/8
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"11.0.0.0", false},
		{"9.255.255.255", false},

		// 100.64.0.0/10 (CGNAT)
		{"100.64.0.1", true},
		{"100.127.255.255", true},
		{"100.63.255.255", false},
		{"100.128.0.0", false},

		// 172.16.0.0/12
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"172.15.255.255", false},
		{"172.32.0.0", false},

		// 192.168.0.0/16
		{"192.168.0.1", true},
		{"192.168.255.255", true},
		{"192.167.255.255", false},
		{"192.169.0.0", false},

		// 192.0.0.0/24 (IETF Protocol Assignments)
		{"192.0.0.1", true},
		{"192.0.0.255", true},
		// 192.0.2.0/24 (TEST-NET-1)
		{"192.0.2.1", true},
		{"192.0.2.255", true},
		// Check between these ranges
		{"192.0.1.1", false},

		// 198.18.0.0/15 (Benchmarking)
		{"198.18.0.1", true},
		{"198.19.255.255", true},
		{"198.17.255.255", false},
		{"198.20.0.0", false},

		// 198.51.100.0/24 (TEST-NET-2)
		{"198.51.100.1", true},
		{"198.51.101.0", false},
		{"198.51.99.0", false},

		// 203.0.113.0/24 (TEST-NET-3)
		{"203.0.113.1", true},
		{"203.0.112.255", false},
		{"203.0.114.0", false},

		// Class E and Broadcast
		{"240.0.0.0", true},
		{"255.255.255.255", true},
		{"224.0.0.0", false}, // Multicast, not strictly private in this function logic?
		// Logic says "if ip[0] >= 240". 224 < 240. So returns false.
		// Multicast is generally not "Private Network" in the sense of RFC1918, but is special.
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			ip4 := ip.To4()
			if ip4 == nil {
				t.Fatalf("Invalid IPv4 test case: %s", tt.ip)
			}
			if got := IsPrivateNetworkIPv4(ip4); got != tt.want {
				t.Errorf("IsPrivateNetworkIPv4(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsPrivateNetworkIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		// IPv4 (Delegates to IsPrivateNetworkIPv4)
		{"10.0.0.1", true},
		{"8.8.8.8", false},

		// Unspecified
		{"0.0.0.0", true},
		{"::", true},

		// Loopback (Explicitly False in IsPrivateNetworkIP)
		{"127.0.0.1", false},
		{"::1", false},

		// IPv6 Unique Local
		{"fc00::1", true},
		{"fd00::1", true},
		{"fbff::1", false}, // Outside fc00::/7

		// IPv6 Documentation
		{"2001:db8::1", true},
		{"2001:db9::1", false},

		// NAT64 + Private IPv4
		{"64:ff9b::10.0.0.1", true},
		{"64:ff9b::c0a8:0101", true}, // 192.168.1.1
		// NAT64 + Public IPv4
		{"64:ff9b::8.8.8.8", false},

		// IPv4 Compatible + Private IPv4
		{"::192.168.1.1", true},
		// IPv4 Compatible + Public IPv4
		{"::8.8.8.8", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if got := IsPrivateNetworkIP(ip); got != tt.want {
				t.Errorf("IsPrivateNetworkIP(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		// IsPrivateNetworkIP cases (should be true)
		{"10.0.0.1", true},
		{"fc00::1", true},
		{"0.0.0.0", true},
		{"::", true},

		// Loopback (should be true here)
		{"127.0.0.1", true},
		{"::1", true},

		// Link-Local IPv4
		{"169.254.1.1", true},
		{"169.255.0.0", false},

		// Link-Local IPv6
		{"fe80::1", true},
		{"feb0::1", true}, // fe80::/10 includes up to febf
		{"fec0::1", false}, // Site local (deprecated) not explicitly checked, falls through

		// IPv4 Compatible Loopback
		{"::127.0.0.1", true},
		// IPv4 Compatible Link-Local
		{"::169.254.1.1", true},

		// NAT64 Loopback
		{"64:ff9b::127.0.0.1", true},
		// NAT64 Link-Local
		{"64:ff9b::169.254.1.1", true},

		// Public
		{"8.8.8.8", false},
		{"2001:4860:4860::8888", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if got := IsPrivateIP(ip); got != tt.want {
				t.Errorf("IsPrivateIP(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}
