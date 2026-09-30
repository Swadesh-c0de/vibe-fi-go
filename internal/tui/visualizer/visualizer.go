package visualizer

import (
	"fmt"
	"strings"
	"vibe-fi/internal/player"
	"vibe-fi/internal/tui/theme"
)

// VisualizerMode defines the active visualizer rendering algorithm.
type VisualizerMode int

const (
	ModeCavaWave VisualizerMode = iota
	ModeNeonFlame
	ModeStereoBars
)

func (m VisualizerMode) String() string {
	switch m {
	case ModeNeonFlame:
		return "NEON FLAME"
	case ModeStereoBars:
		return "STEREO BARS"
	default:
		return "CAVA WAVE"
	}
}

var (
	BlockChars = []string{" ", " ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	PeakChar   = "▔"
)

// CachedBlockChars holds pre-rendered ANSI strings for block elements in a specific theme.
type CachedBlockChars struct {
	ThemeName    string
	Base         [9]string
	Mid          [9]string
	High         [9]string
	Peak         string
	PeakSparkle  string
	PeakTriangle string
	FullBase     string
	FullMid      string
	FullHigh     string
}

// Visualizer maintains physics, animation history, and renders visualizer frames.
type Visualizer struct {
	CurrentProfile TrackVisualProfile

	// Cava Wave state
	cavaBars  []float32
	cavaPeaks []float32
	cavaHold  []int
	cavaFall  []float32

	// Neon Flame state
	flameBars  []float32
	flamePeaks []float32
	flameHold  []int
	flameFall  []float32

	// Stereo Bars state
	stereoBars  []float32
	stereoPeaks []float32
	stereoHold  []int
	stereoFall  []float32

	fetchAnimFrame int

	// Reusable rendering buffers (Zero-allocation render loop)
	cachedBlocks CachedBlockChars
	gridW        int
	gridH        int
	grid         [][]string
	lines        []string
	targets      []float32
	smoothed     []float32
}

// NewVisualizer creates a new visualizer engine instance.
func NewVisualizer() *Visualizer {
	v := &Visualizer{}
	v.Reset()
	return v
}

// Reset clears all animation and peak vectors.
func (v *Visualizer) Reset() {
	v.cavaBars = nil
	v.cavaPeaks = nil
	v.cavaHold = nil
	v.cavaFall = nil

	v.flameBars = nil
	v.flamePeaks = nil
	v.flameHold = nil
	v.flameFall = nil

	v.stereoBars = nil
	v.stereoPeaks = nil
	v.stereoHold = nil
	v.stereoFall = nil
}

// EnsureCachedBlocks pre-renders styled block glyphs for the active theme, eliminating per-cell styling.
func (v *Visualizer) EnsureCachedBlocks(styles theme.Styles) {
	if v.cachedBlocks.ThemeName == styles.Theme.Name && v.cachedBlocks.ThemeName != "" {
		return
	}
	v.cachedBlocks.ThemeName = styles.Theme.Name
	for i := 0; i < 9; i++ {
		char := BlockChars[i]
		v.cachedBlocks.Base[i] = styles.VizBase.Render(char)
		v.cachedBlocks.Mid[i] = styles.VizMid.Render(char)
		v.cachedBlocks.High[i] = styles.VizHigh.Render(char)
	}
	v.cachedBlocks.Peak = styles.VizPeak.Render(PeakChar)
	v.cachedBlocks.PeakSparkle = styles.VizPeak.Render("✦")
	v.cachedBlocks.PeakTriangle = styles.VizPeak.Render("▲")
	v.cachedBlocks.FullBase = v.cachedBlocks.Base[8]
	v.cachedBlocks.FullMid = v.cachedBlocks.Mid[8]
	v.cachedBlocks.FullHigh = v.cachedBlocks.High[8]
}

// PrepareGrid returns the reusable 2D grid reset to spaces without heap allocations.
func (v *Visualizer) PrepareGrid(drawW, drawH int) [][]string {
	if v.gridH != drawH || v.gridW != drawW || len(v.grid) != drawH {
		v.gridH = drawH
		v.gridW = drawW
		v.grid = make([][]string, drawH)
		for y := 0; y < drawH; y++ {
			v.grid[y] = make([]string, drawW)
		}
		v.lines = make([]string, drawH)
	}
	for y := 0; y < drawH; y++ {
		row := v.grid[y]
		for x := 0; x < drawW; x++ {
			row[x] = " "
		}
	}
	return v.grid
}

// PrepareBuffers returns reusable float slices for target and smoothed spectrums.
func (v *Visualizer) PrepareBuffers(numBars int) ([]float32, []float32) {
	if len(v.targets) != numBars {
		v.targets = make([]float32, numBars)
		v.smoothed = make([]float32, numBars)
	}
	return v.targets, v.smoothed
}

// BuildLines joins the grid rows into the final output slice reusing v.lines.
func (v *Visualizer) BuildLines(drawH int) []string {
	for y := 0; y < drawH; y++ {
		v.lines[y] = strings.Join(v.grid[y], "")
	}
	return v.lines
}

// RenderHeader computes the visualizer box title bar (including metronome & peak).
func (v *Visualizer) RenderHeader(p player.AudioPlayer, mode VisualizerMode) string {
	if p.IsLoading() || p.IsBuffering() {
		state := "FETCHING STREAM"
		if p.IsBuffering() {
			state = "BUFFERING"
		}
		return fmt.Sprintf("VISUALIZER: %s [%s]", mode.String(), state)
	}

	if p.IsIdle() {
		return fmt.Sprintf("VISUALIZER: %s [IDLE]", mode.String())
	}

	modeTitle := fmt.Sprintf("VISUALIZER: %s", mode.String())
	isActive := p.IsPlaying() && !p.IsPaused() && !p.IsIdle() && !p.IsLoading()

	if isActive {
		stats := p.GetAudioStats()
		pos := p.Position()
		trackBPS := v.CurrentProfile.BPM / 60.0

		beatStep := 0
		if trackBPS > 0.05 {
			beatStep = int(pos*float64(trackBPS)) % 4
		}
		metro := "[♫"
		for b := 0; b < 4; b++ {
			if b == beatStep {
				metro += " ●"
			} else {
				metro += " ○"
			}
		}
		metro += " ]"

		if stats.Valid && stats.PeakOverall > 0.001 {
			lvlPct := int(stats.PeakOverall * 100.0)
			modeTitle += fmt.Sprintf(" %s [%d%% PEAK]", metro, lvlPct)
		} else {
			bpmDisplay := int(v.CurrentProfile.BPM)
			modeTitle += fmt.Sprintf(" %s [~%d BPM]", metro, bpmDisplay)
		}
	} else if p.IsPaused() {
		modeTitle += " [PAUSED]"
	}

	return modeTitle
}

// RenderBody produces the inner lines of the visualizer (height x width).
func (v *Visualizer) RenderBody(drawW, drawH int, p player.AudioPlayer, mode VisualizerMode, styles theme.Styles) []string {
	if drawW <= 0 || drawH <= 0 {
		return nil
	}

	v.CurrentProfile = ComputeProfile(p)

	// Case 1: Loading or Buffering
	if p.IsLoading() || p.IsBuffering() {
		v.Reset()
		v.fetchAnimFrame++
		dotStep := (v.fetchAnimFrame / 6) % 4
		pulse := "["
		for b := 0; b < 4; b++ {
			if b == dotStep {
				pulse += " ●"
			} else {
				pulse += " ○"
			}
		}
		pulse += " ]"

		msg := ":: Fetching audio stream..."
		if p.IsBuffering() {
			msg = ":: Buffering audio stream..."
		}

		lines := make([]string, drawH)
		msgY := drawH / 2
		if msgY > 0 {
			msgY--
		}
		pulseY := msgY + 2
		if pulseY >= drawH {
			pulseY = drawH - 1
		}

		for y := 0; y < drawH; y++ {
			if y == msgY {
				leftPad := (drawW - len(msg)) / 2
				if leftPad < 0 {
					leftPad = 0
				}
				lines[y] = strings.Repeat(" ", leftPad) + styles.VizMid.Render(msg)
			} else if y == pulseY {
				leftPad := (drawW - len(pulse)) / 2
				if leftPad < 0 {
					leftPad = 0
				}
				lines[y] = strings.Repeat(" ", leftPad) + styles.ProgressBar.Render(pulse)
			} else {
				lines[y] = strings.Repeat(" ", drawW)
			}
		}
		return lines
	}

	// Case 2: Idle
	if p.IsIdle() {
		v.Reset()
		idleMsg := ":: No audio playing. Press [S] to search or [L] for library."
		if len(idleMsg) > drawW-4 {
			idleMsg = ":: Press [S] to search or [L] for library."
		}
		leftPad := (drawW - len(idleMsg)) / 2
		if leftPad < 0 {
			leftPad = 0
		}

		lines := make([]string, drawH)
		midY := drawH / 2
		for y := 0; y < drawH; y++ {
			if y == midY {
				lines[y] = strings.Repeat(" ", leftPad) + styles.StatusDim.Render(idleMsg)
			} else {
				lines[y] = strings.Repeat(" ", drawW)
			}
		}
		return lines
	}

	// Case 3: Active Playback (or Paused)
	pos := p.Position()
	stats := p.GetAudioStats()
	vol := float32(p.Volume()) / 100.0
	if vol < 0.2 {
		vol = 0.2
	}
	if vol > 1.2 {
		vol = 1.2
	}

	v.EnsureCachedBlocks(styles)

	switch mode {
	case ModeNeonFlame:
		return v.renderNeonFlame(drawW, drawH, p, pos, vol, stats, styles)
	case ModeStereoBars:
		return v.renderStereoBars(drawW, drawH, p, pos, vol, stats, styles)
	default:
		return v.renderCavaWave(drawW, drawH, p, pos, vol, stats, styles)
	}
}
