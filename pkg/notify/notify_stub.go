//go:build !windows

package notify

import "errors"

var ErrUnsupported = errors.New("notifications not supported")

// Send is a no-op on unsupported platforms.
func Send(title, message string) error {
	return ErrUnsupported
}
