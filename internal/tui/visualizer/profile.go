package visualizer

import (
	"strings"
	"vibe-fi/internal/player"
)

// TrackVisualProfile produces a deterministic acoustic profile per song.
type TrackVisualProfile struct {
	TrackID        string
	BPM            float32
	BassWeight     float32
	MidWeight      float32
	TrebleWeight   float32
	RhythmSwing    float32
	EnergyVariance float32
}

// ComputeProfile calculates musical groove parameters matching C++ implementation.
func ComputeProfile(p player.AudioPlayer) TrackVisualProfile {
	title := p.GetMetadata("media-title")
	if title == "" {
		title = p.GetMetadata("filename")
	}
	if title == "" {
		title = "default"
	}

	var hash uint64 = 5381
	for _, c := range title {
		hash = ((hash << 5) + hash) + uint64(c)
	}

	lower := strings.ToLower(title)
	prof := TrackVisualProfile{TrackID: title}

	if strings.Contains(lower, "lofi") || strings.Contains(lower, "chill") ||
		strings.Contains(lower, "slow") || strings.Contains(lower, "ambient") ||
		strings.Contains(lower, "sleep") {
		prof.BPM = 74.0 + float32(hash%16)
		prof.BassWeight = 1.35
		prof.MidWeight = 0.90
		prof.TrebleWeight = 0.70
	} else if strings.Contains(lower, "remix") || strings.Contains(lower, "club") ||
		strings.Contains(lower, "dance") || strings.Contains(lower, "edm") ||
		strings.Contains(lower, "house") || strings.Contains(lower, "bass") {
		prof.BPM = 124.0 + float32(hash%18)
		prof.BassWeight = 1.45
		prof.MidWeight = 1.05
		prof.TrebleWeight = 1.30
	} else if strings.Contains(lower, "rock") || strings.Contains(lower, "metal") ||
		strings.Contains(lower, "punk") || strings.Contains(lower, "guitar") {
		prof.BPM = 132.0 + float32(hash%32)
		prof.BassWeight = 1.10
		prof.MidWeight = 1.40
		prof.TrebleWeight = 1.20
	} else if strings.Contains(lower, "rap") || strings.Contains(lower, "hip hop") ||
		strings.Contains(lower, "trap") || strings.Contains(lower, "drill") {
		prof.BPM = 88.0 + float32(hash%24)
		prof.BassWeight = 1.55
		prof.MidWeight = 1.00
		prof.TrebleWeight = 1.10
	} else {
		prof.BPM = 85.0 + float32(hash%60)
		prof.BassWeight = 0.85 + float32((hash>>4)%50)*0.01
		prof.MidWeight = 0.85 + float32((hash>>8)%40)*0.01
		prof.TrebleWeight = 0.80 + float32((hash>>12)%45)*0.01
	}

	prof.RhythmSwing = float32((hash>>16)%25) * 0.01
	prof.EnergyVariance = 0.88 + float32((hash>>20)%24)*0.01
	return prof
}
