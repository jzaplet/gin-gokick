package clientinfo

import (
	"net/netip"
	"strings"
)

const MaxUserAgent = 512

type Client struct {
	UserAgent string

	IP string
}

func UserAgent(userAgent string) string {
	clean := []rune(strings.ReplaceAll(strings.ToValidUTF8(userAgent, ""), "\x00", ""))

	return string(clean[:min(len(clean), MaxUserAgent)])
}

func Address(ip string) *netip.Addr {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return nil
	}

	addr = addr.Unmap()

	return &addr
}

func SameAddress(stored, seen *netip.Addr) bool {
	if seen == nil {
		return true
	}

	return stored != nil && *stored == *seen
}
