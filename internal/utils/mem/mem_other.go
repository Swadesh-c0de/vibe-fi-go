//go:build !linux

package mem

import "runtime/debug"

// TuneMemory configures GC target and memory limits on non-Linux platforms.
func TuneMemory() {
	debug.SetGCPercent(40)
	debug.SetMemoryLimit(50 * 1024 * 1024)
}

// TrimMemory is a no-op on non-glibc platforms.
func TrimMemory() {}

// PeriodicTrim flushes Go runtime scavenged pages to the OS.
func PeriodicTrim() {
	debug.FreeOSMemory()
}
