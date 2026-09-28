package visualizer

import (
	"math"
	"math/rand"
	"strings"
	"vibe-fi/internal/player"
	"vibe-fi/internal/tui/theme"
)

func (v *Visualizer) renderNeonFlame(drawW, drawH int, p player.AudioPlayer, pos float64, vol float32, stats player.AudioLevelStats, styles theme.Styles) []string {
	isActive := p.IsPlaying() && !p.IsPaused() && !p.IsIdle() && !p.IsLoading()
	maxSubLevels := float32(drawH * 8)

	liveRMS := float32(0.35)
	livePeak := float32(0.45)
	liveLeft := liveRMS
	liveRight := liveRMS
	livePitch := float32(1.0)

	if stats.Valid {
		liveRMS = stats.RMSOverall
		livePeak = stats.PeakOverall
		liveLeft = stats.RMSLeft
		liveRight = stats.RMSRight
		livePitch = stats.ZeroCrossings * 12.0
		if livePitch < 0.20 {
			livePitch = 0.20
		}
		if livePitch > 1.50 {
			livePitch = 1.50
		}
	}

	attackSpeed := float32(0.90)
	decaySpeed := float32(0.26)

	barW := 1
	if drawW >= 80 {
		barW = 2
	}
	gap := 1
	slotW := barW + gap
	numBars := drawW / slotW
	if numBars < 6 {
		numBars = 6
	}
	totalBarsW := numBars*slotW - gap
	startX := (drawW - totalBarsW) / 2
	if startX < 0 {
		startX = 0
	}

	if len(v.flameBars) != numBars {
		v.flameBars = make([]float32, numBars)
		v.flamePeaks = make([]float32, numBars)
		v.flameHold = make([]int, numBars)
		v.flameFall = make([]float32, numBars)
	}

	if isActive {
		trackBPS := v.CurrentProfile.BPM / 60.0
		beatInterval := float64(1.0 / math.Max(0.1, float64(trackBPS)))
		centerIdx := numBars / 2

		kickPhase := math.Mod(pos, beatInterval) / beatInterval
		kick := float32(math.Exp(-kickPhase*(5.0+float64(v.CurrentProfile.RhythmSwing)*3.0))) * v.CurrentProfile.BassWeight

		snarePhase := math.Mod(pos+beatInterval*0.5, beatInterval) / beatInterval
		snare := float32(math.Exp(-snarePhase*7.0)) * v.CurrentProfile.MidWeight

		hihatPhase := math.Mod(pos, beatInterval*0.25) / (beatInterval * 0.25)
		hihat := float32(math.Exp(-hihatPhase*11.0)) * v.CurrentProfile.TrebleWeight

		for i := 0; i < numBars; i++ {
			div := centerIdx
			if div < 1 {
				div = 1
			}
			dist := float32(math.Abs(float64(i-centerIdx))) / float32(div)
			channelEnergy := liveLeft
			if i >= centerIdx {
				channelEnergy = liveRight
			}
			if !stats.Valid {
				channelEnergy = liveRMS
			}

			centerPunch := kick * float32(math.Max(0.0, float64(1.0-dist*2.6))) * (0.45 + livePeak*0.85)
			midDance := snare * float32(math.Max(0.0, float64(1.0-float32(math.Abs(float64(dist-0.45)))*3.0))) * (0.35 + channelEnergy*0.85)
			trebleFlicker := (hihat*0.40 + float32(rand.Intn(16))*0.01) * float32(math.Max(0.0, float64(dist-0.35))) * 2.2

			w1 := float32(math.Sin(pos*(float64(trackBPS)*3.8*float64(livePitch))+(float64(i)*0.42))) * 0.22
			w2 := float32(math.Cos(pos*(float64(trackBPS)*7.2*float64(livePitch))-(float64(dist)*4.8))) * 0.16
			w3 := float32(math.Sin(pos*(float64(trackBPS)*1.5)+(float64(dist)*math.Pi))) * 0.12
			flameFlutter := w1 + w2 + w3

			baseTilt := (1.05 - dist*0.38) * (0.36 + flameFlutter)
			rawEnergy := (baseTilt*(0.30+channelEnergy*0.70) + centerPunch*0.85 + midDance*0.65 + trebleFlicker) * vol

			if stats.Valid {
				rawEnergy *= (0.40 + livePeak*0.80)
			}
			if rawEnergy < 0.04 {
				rawEnergy = 0.04
			}
			if rawEnergy > 1.0 {
				rawEnergy = 1.0
			}

			target := rawEnergy * maxSubLevels
			if target > v.flameBars[i] {
				v.flameBars[i] += (target - v.flameBars[i]) * attackSpeed
			} else {
				v.flameBars[i] -= (v.flameBars[i] - target) * decaySpeed
			}

			if v.flameBars[i] >= v.flamePeaks[i] {
				v.flamePeaks[i] = v.flameBars[i]
				v.flameHold[i] = 3
				v.flameFall[i] = 0.0
			} else {
				if v.flameHold[i] > 0 {
					v.flameHold[i]--
				} else {
					v.flameFall[i] += 0.65
					v.flamePeaks[i] -= v.flameFall[i]
					if v.flamePeaks[i] < v.flameBars[i] {
						v.flamePeaks[i] = v.flameBars[i]
					}
				}
			}
		}
	} else {
		for i := 0; i < numBars; i++ {
			v.flameBars[i] *= 0.85
			v.flamePeaks[i] *= 0.85
			if v.flameBars[i] < 0.5 {
				v.flameBars[i] = 0.0
			}
			if v.flamePeaks[i] < 0.5 {
				v.flamePeaks[i] = 0.0
			}
		}
	}

	grid := make([][]string, drawH)
	for y := 0; y < drawH; y++ {
		grid[y] = make([]string, drawW)
		for x := 0; x < drawW; x++ {
			grid[y][x] = " "
		}
	}

	for i := 0; i < numBars; i++ {
		val := int(v.flameBars[i])
		fullCells := val / 8
		rem := val % 8
		peakCell := int(v.flamePeaks[i]) / 8
		barX := startX + i*slotW

		for y := 0; y < drawH; y++ {
			drawY := drawH - 1 - y
			style := styles.VizBase
			if y >= drawH*2/3 {
				style = styles.VizHigh
			} else if y >= drawH/3 {
				style = styles.VizMid
			}

			if y < fullCells {
				for k := 0; k < barW && (barX+k) < drawW; k++ {
					grid[drawY][barX+k] = style.Render("█")
				}
			} else if y == fullCells && rem > 0 {
				char := BlockChars[rem]
				for k := 0; k < barW && (barX+k) < drawW; k++ {
					grid[drawY][barX+k] = style.Render(char)
				}
			} else if y == peakCell && peakCell > fullCells && peakCell < drawH {
				crownSym := PeakChar
				if peakCell >= drawH*2/3 {
					crownSym = "✦"
				} else if peakCell >= drawH/3 {
					crownSym = "▲"
				}
				for k := 0; k < barW && (barX+k) < drawW; k++ {
					grid[drawY][barX+k] = styles.VizPeak.Render(crownSym)
				}
			}
		}
	}

	lines := make([]string, drawH)
	for y := 0; y < drawH; y++ {
		lines[y] = strings.Join(grid[y], "")
	}
	return lines
}
