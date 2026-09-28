package net

import (
	"net"
	"time"
)

// IsOnline checks network connectivity by connecting to a public DNS port.
func IsOnline() bool {
	timeout := 1500 * time.Millisecond
	conn, err := net.DialTimeout("tcp", "1.1.1.1:53", timeout)
	if err != nil {
		conn, err = net.DialTimeout("tcp", "8.8.8.8:53", timeout)
	}
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
