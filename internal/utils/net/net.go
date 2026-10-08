package net

import (
	"net"
	"sync"
	"time"
)

var (
	netMu        sync.RWMutex
	cachedOnline bool
	lastChecked  time.Time
)

// IsOnline checks network connectivity by connecting to a public DNS port.
// Results are cached for 3 seconds when online (1s when offline) to eliminate
// redundant TCP network handshakes on rapid lookups.
func IsOnline() bool {
	netMu.RLock()
	ttl := 3 * time.Second
	if !cachedOnline {
		ttl = 1 * time.Second
	}
	if time.Since(lastChecked) < ttl {
		online := cachedOnline
		netMu.RUnlock()
		return online
	}
	netMu.RUnlock()

	netMu.Lock()
	defer netMu.Unlock()

	// Double-check under lock
	if time.Since(lastChecked) < ttl {
		return cachedOnline
	}

	timeout := 1500 * time.Millisecond
	conn, err := net.DialTimeout("tcp", "1.1.1.1:53", timeout)
	if err != nil {
		conn, err = net.DialTimeout("tcp", "8.8.8.8:53", timeout)
	}

	lastChecked = time.Now()
	if err != nil {
		cachedOnline = false
		return false
	}
	_ = conn.Close()
	cachedOnline = true
	return true
}

