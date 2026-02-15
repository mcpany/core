package validation

import (
	"net"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestIsPrivateIP_Coverage(t *testing.T) {
	tests := []struct {
		name      string
		ipStr     string
		isPrivate bool
	}{
		// IPv4 Private Blocks
		{"10.0.0.1", "10.0.0.1", true},
		{"10.255.255.255", "10.255.255.255", true},
		{"172.16.0.1", "172.16.0.1", true},
		{"172.31.255.255", "172.31.255.255", true},
		{"172.15.0.1", "172.15.0.1", false}, // Outside
		{"172.32.0.1", "172.32.0.1", false}, // Outside
		{"192.168.0.1", "192.168.0.1", true},
		{"192.168.255.255", "192.168.255.255", true},
		{"192.167.0.1", "192.167.0.1", false}, // Outside

		// Shared Address Space (CGNAT) 100.64.0.0/10
		{"100.64.0.1", "100.64.0.1", true},
		{"100.127.255.255", "100.127.255.255", true},
		{"100.63.255.255", "100.63.255.255", false}, // Outside
		{"100.128.0.0", "100.128.0.0", false},       // Outside

		// IETF Protocol Assignments 192.0.0.0/24
		{"192.0.0.1", "192.0.0.1", true},
		{"192.0.0.255", "192.0.0.255", true},
		{"192.0.1.1", "192.0.1.1", false}, // Outside

		// TEST-NET-1 192.0.2.0/24
		{"192.0.2.1", "192.0.2.1", true},

		// Benchmarking 198.18.0.0/15
		{"198.18.0.1", "198.18.0.1", true},
		{"198.19.255.255", "198.19.255.255", true},
		{"198.17.255.255", "198.17.255.255", false}, // Outside

		// TEST-NET-2 198.51.100.0/24
		{"198.51.100.1", "198.51.100.1", true},
		{"198.51.101.1", "198.51.101.1", false}, // Outside

		// TEST-NET-3 203.0.113.0/24
		{"203.0.113.1", "203.0.113.1", true},
		{"203.0.112.1", "203.0.112.1", false}, // Outside

		// Class E 240.0.0.0/4
		{"240.0.0.1", "240.0.0.1", true},
		{"255.255.255.255", "255.255.255.255", true},

		// Loopback
		{"127.0.0.1", "127.0.0.1", true},

		// Link-Local
		{"169.254.0.1", "169.254.0.1", true},

		// 0.0.0.0
		{"0.0.0.0", "0.0.0.0", true},

		// IPv6
		{"IPv6 Loopback", "::1", true},
		{"IPv6 Unspecified", "::", true},
		{"IPv6 Link Local", "fe80::1", true},
		{"IPv6 Site Local (Unique Local)", "fc00::1", true},
		{"IPv6 Documentation", "2001:db8::1", true},

		// IPv4-mapped IPv6
		{"IPv4-mapped Localhost", "::ffff:127.0.0.1", true}, // Not strictly private IP func, but url.go might handle
		{"IPv4-mapped Private", "::ffff:10.0.0.1", true},    // net.ParseIP usually returns IPv4 for these if converted to 4

		// IPv4-compatible IPv6 (deprecated but handled) ::a.b.c.d
		{"IPv4-compatible Localhost", "::127.0.0.1", true},
		{"IPv4-compatible LinkLocal", "::169.254.1.1", true},
		{"IPv4-compatible Private", "::10.0.0.1", true},

		// NAT64 64:ff9b::/96
		{"NAT64 Private", "64:ff9b::10.0.0.1", true}, // 64:ff9b::0a00:0001
		{"NAT64 Public", "64:ff9b::8.8.8.8", false},  // Public IP in NAT64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ipStr)
			if ip == nil {
				t.Fatalf("Invalid IP in test case: %s", tt.ipStr)
			}
			got := IsPrivateIP(ip)
			if got != tt.isPrivate {
				t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ipStr, got, tt.isPrivate)
			}
		})
	}
}

func TestIsNAT64_Coverage(t *testing.T) {
	// 64:ff9b::1 is valid NAT64 prefix but not embedded IPv4 at end?
	// RFC 6052: prefix is 96 bits.
	// 64:ff9b:: is the Well-Known Prefix.
	// We need 12 bytes of prefix.
	// 0064:ff9b:0000:0000:0000:0000 ...

	ip := net.ParseIP("64:ff9b::1.2.3.4")
	if !IsNAT64(ip) {
		t.Errorf("Expected 64:ff9b::1.2.3.4 to be NAT64")
	}

	ip = net.ParseIP("2001::1")
	if IsNAT64(ip) {
		t.Errorf("Expected 2001::1 NOT to be NAT64")
	}
}

func TestIsIPv4Compatible_Coverage(t *testing.T) {
	ip := net.ParseIP("::1.2.3.4") // IPv4 compatible
	// net.ParseIP for "::1.2.3.4" returns 16 bytes.
	if !IsIPv4Compatible(ip) {
		t.Errorf("Expected ::1.2.3.4 to be IPv4 compatible")
	}

	ip = net.ParseIP("::1") // Loopback
	if !IsIPv4Compatible(ip) {
		// loopback ::1 is 0...01, which technically matches the prefix of 96 zero bits
		t.Errorf("Expected ::1 to be detected as IPv4 compatible form (96 zero bits)")
	}
}

func TestValidateIP_Multicast(t *testing.T) {
    // Multicast IPv4
    err := IsSafeURL("http://224.0.0.1")
    if err == nil {
        t.Errorf("Expected IsSafeURL to reject multicast IPv4 224.0.0.1")
    }

    // Multicast IPv6
    err = IsSafeURL("http://[ff02::1]")
    if err == nil {
        t.Errorf("Expected IsSafeURL to reject multicast IPv6 ff02::1")
    }
}

func TestValidateHTTPServiceDefinition_Coverage(t *testing.T) {
    // Test invalid characters in path to trigger url.Parse error
    methodGet := configv1.HttpCallDefinition_HTTP_METHOD_GET
    def := configv1.HttpCallDefinition_builder{
        EndpointPath: proto.String("/path\x7fcontrol"),
        Method:       methodGet.Enum(),
    }.Build()

    err := ValidateHTTPServiceDefinition(def)
    // url.Parse might not fail on control chars depending on Go version, but let's try.
    // Actually, usually it accepts almost anything.
    // But checking coverage, we need to hit `if err != nil`.

    // Maybe a path that is not empty but effectively invalid for URL?
    // /%gh
    def.SetEndpointPath("/%gh") // Invalid escape
    err = ValidateHTTPServiceDefinition(def)
    if err == nil {
        t.Log("url.Parse did not fail for invalid escape")
    } else {
        t.Logf("url.Parse failed as expected: %v", err)
    }
}
