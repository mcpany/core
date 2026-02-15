package validation

import (
	"net"
)

var privateNetworkBlocksIPv6 []*net.IPNet

func init() {
	// RFC4193 + others
	// Note: IPv4 blocks are handled by isPrivateNetworkIPv4 fast path, so we only need IPv6 here.
	for _, cidr := range []string{
		"fc00::/7",      // RFC4193 unique local address
		"2001:db8::/32", // IPv6 documentation (RFC 3849)
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err == nil {
			privateNetworkBlocksIPv6 = append(privateNetworkBlocksIPv6, block)
		}
	}
}

// IsPrivateNetworkIP checks if the IP address is a private network address.
// This includes RFC1918, RFC4193 (Unique Local), and RFC6598 (CGNAT).
// It does NOT include loopback or link-local addresses.
//
// Parameters:
//   - ip: The IP address to check.
//
// Returns:
//   - bool: True if the IP is a private network address, false otherwise.
func IsPrivateNetworkIP(ip net.IP) bool {
	// Treat unspecified addresses (0.0.0.0 and ::) as private.
	// 0.0.0.0 is also covered by isPrivateNetworkIPv4, but :: wasn't.
	if ip.IsUnspecified() {
		return true
	}

	if ip.IsLoopback() {
		return false
	}

	if ip4 := ip.To4(); ip4 != nil {
		// IPv4 fast path: check directly to avoid linear scan of net.IPNet slices
		return IsPrivateNetworkIPv4(ip4)
	}

	if IsNAT64(ip) || IsIPv4Compatible(ip) {
		// Last 4 bytes are the IPv4 address
		ip4 := ip[12:16]
		return IsPrivateNetworkIPv4(ip4)
	}

	for _, block := range privateNetworkBlocksIPv6 {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// IsNAT64 checks for NAT64 (IPv4-embedded IPv6) addresses - 64:ff9b::/96 (RFC 6052).
//
// Parameters:
//   - ip: The IP address to check.
//
// Returns:
//   - bool: True if the IP is a NAT64 address, false otherwise.
func IsNAT64(ip net.IP) bool {
	// 64:ff9b:: expands to 0064:ff9b:0000:0000:0000:0000 (96 bits)
	return len(ip) == net.IPv6len &&
		ip[0] == 0x00 && ip[1] == 0x64 && ip[2] == 0xff && ip[3] == 0x9b &&
		ip[4] == 0 && ip[5] == 0 && ip[6] == 0 && ip[7] == 0 &&
		ip[8] == 0 && ip[9] == 0 && ip[10] == 0 && ip[11] == 0
}

// IsIPv4Compatible checks for IPv4-compatible IPv6 addresses (::a.b.c.d).
//
// Parameters:
//   - ip: The IP address to check.
//
// Returns:
//   - bool: True if the IP is an IPv4-compatible IPv6 address, false otherwise.
func IsIPv4Compatible(ip net.IP) bool {
	// First 12 bytes are 0.
	return len(ip) == net.IPv6len &&
		ip[0] == 0 && ip[1] == 0 && ip[2] == 0 && ip[3] == 0 &&
		ip[4] == 0 && ip[5] == 0 && ip[6] == 0 && ip[7] == 0 &&
		ip[8] == 0 && ip[9] == 0 && ip[10] == 0 && ip[11] == 0
}

// IsNAT64LinkLocal checks if a NAT64 address embeds a link-local IPv4 address.
//
// Parameters:
//   - ip: The IP address to check.
//
// Returns:
//   - bool: True if the IP is a NAT64 link-local address, false otherwise.
func IsNAT64LinkLocal(ip net.IP) bool {
	if !IsNAT64(ip) {
		return false
	}
	// Extract embedded IPv4 (last 4 bytes)
	ip4 := ip[12:16]
	// Check for Link-local (169.254.0.0/16)
	return ip4[0] == 169 && ip4[1] == 254
}

// IsNAT64Loopback checks if a NAT64 address embeds a loopback IPv4 address.
//
// Parameters:
//   - ip: The IP address to check.
//
// Returns:
//   - bool: True if the IP is a NAT64 loopback address, false otherwise.
func IsNAT64Loopback(ip net.IP) bool {
	if !IsNAT64(ip) {
		return false
	}
	// Extract embedded IPv4 (last 4 bytes)
	ip4 := ip[12:16]
	// Check for Loopback (127.0.0.0/8)
	return ip4[0] == 127
}

// IsPrivateIP checks if the IP address is a private, link-local, or loopback address.
//
// Parameters:
//   - ip: The IP address to check.
//
// Returns:
//   - bool: True if the IP is private, link-local, or loopback, false otherwise.
func IsPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsUnspecified() {
		return true
	}

	if ip4 := ip.To4(); ip4 != nil {
		// Link-local (169.254.0.0/16)
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
		return IsPrivateNetworkIPv4(ip4)
	}

	// IPv6 Link-local (fe80::/10)
	if len(ip) == net.IPv6len && ip[0] == 0xfe && ip[1]&0xc0 == 0x80 {
		return true
	}

	// Check for IPv4-compatible IPv6 addresses (::a.b.c.d) for Loopback/Link-local
	if IsIPv4Compatible(ip) {
		ip4 := ip[12:16]
		// Loopback (127.0.0.0/8)
		if ip4[0] == 127 {
			return true
		}
		// Link-local (169.254.0.0/16)
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
	}

	if IsNAT64Loopback(ip) || IsNAT64LinkLocal(ip) {
		return true
	}

	return IsPrivateNetworkIP(ip)
}

// IsPrivateNetworkIPv4 checks if an IPv4 address is private.
// ip must be a valid 4-byte IPv4 address slice.
//
// Parameters:
//   - ip: The IPv4 address slice to check.
//
// Returns:
//   - bool: True if the IP is private, false otherwise.
func IsPrivateNetworkIPv4(ip net.IP) bool {
	switch ip[0] {
	case 0:
		return true // 0.0.0.0/8
	case 10:
		return true // 10.0.0.0/8
	case 100:
		return ip[1] >= 64 && ip[1] <= 127 // 100.64.0.0/10
	case 172:
		return ip[1] >= 16 && ip[1] <= 31 // 172.16.0.0/12
	case 192:
		if ip[1] == 168 {
			return true // 192.168.0.0/16
		}
		if ip[1] == 0 {
			return ip[2] == 0 || ip[2] == 2 // 192.0.0.0/24 or 192.0.2.0/24
		}
	case 198:
		if ip[1] == 18 || ip[1] == 19 {
			return true // 198.18.0.0/15
		}
		return ip[1] == 51 && ip[2] == 100 // 198.51.100.0/24
	case 203:
		return ip[1] == 0 && ip[2] == 113 // 203.0.113.0/24
	}

	// Class E (240.0.0.0/4) and Broadcast (255.255.255.255)
	if ip[0] >= 240 {
		return true
	}

	return false
}
