package net

import (
	"testing"
)

func TestIsOnline(t *testing.T) {
	// Call IsOnline twice to test both initial check and cached hit
	res1 := IsOnline()
	res2 := IsOnline()
	if res1 != res2 {
		t.Errorf("expected consistent cached result, got %v and %v", res1, res2)
	}
}
