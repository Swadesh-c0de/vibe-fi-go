//go:build linux

package mem

/*
#include <malloc.h>
#include <stdlib.h>

static inline void c_tune_memory() {
#if defined(__linux__)
    mallopt(M_ARENA_MAX, 2);
    mallopt(M_TRIM_THRESHOLD, 64 * 1024);
    mallopt(M_MMAP_THRESHOLD, 64 * 1024);
#endif
}

static inline void c_trim_memory() {
#if defined(__linux__)
    malloc_trim(0);
#endif
}
*/
import "C"
import "runtime/debug"

// TuneMemory applies glibc arena and mmap tuning to restrict virtual memory expansion,
// and sets GC target pacer and soft memory limit for low-RAM audio playback.
func TuneMemory() {
	C.c_tune_memory()
	debug.SetGCPercent(40)
	debug.SetMemoryLimit(50 * 1024 * 1024) // 50MB soft limit
}

// TrimMemory calls malloc_trim(0) to release glibc heap arenas back to Linux.
func TrimMemory() {
	C.c_trim_memory()
}

// PeriodicTrim executes glibc arena trim and flushes Go runtime scavenged pages to the OS.
func PeriodicTrim() {
	C.c_trim_memory()
	debug.FreeOSMemory()
}
