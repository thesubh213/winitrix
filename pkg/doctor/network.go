package doctor

import (
	"context"
	"net"
	"time"
)

// CheckInternetConnection attempts to connect to a highly available external service to verify internet access.
func CheckInternetConnection(ctx context.Context) bool {
	// Use a dialer with a timeout
	dialer := net.Dialer{
		Timeout: 2 * time.Second,
	}

	// Try connecting to Cloudflare's 1.1.1.1 DNS over TCP port 53
	conn, err := dialer.DialContext(ctx, "tcp", "1.1.1.1:53")
	if err != nil {
		// Fallback to Google DNS just in case 1.1.1.1 is specifically blocked
		conn, err = dialer.DialContext(ctx, "tcp", "8.8.8.8:53")
		if err != nil {
			return false
		}
	}
	if conn != nil {
		_ = conn.Close()
		return true
	}
	return false
}
