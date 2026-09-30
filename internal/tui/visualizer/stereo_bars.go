package visualizer

import (
	"math"
	"math/rand"
	"vibe-fi/internal/player"
	"vibe-fi/internal/tui/theme"
)

func (v *Visualizer) renderStereoBars(drawW, drawH int, p player.AudioPlayer, pos float64, vol float32, stats player.AudioLevelStats, styles theme.Styles) []string {
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
		if livePitch < 0.15 {
			livePitch = 0.15
		}
		if livePitch > 1.60 {
			livePitch = 1.60
		}
	}

	attackSpeed := float32(0.60)
	decaySpeed := float32(0.18)
	if stats.Valid {
		attackSpeed = 0.22 + livePeak*0.55
		decaySpeed = 0.07 + liveRMS*0.14
	}

	barW := 1
	if drawW >= 80 {
		barW = 2
	}
	gap := 1
	slotW := barW + gap
	numBars := drawW / slotW
	if numBars < 4 {
		numBars = 4
	}
	totalBarsW := numBars*slotW - gap
	startX := (drawW - totalBarsW) / 2
	if startX < 0 {
		startX = 0
	}

	if len(v.stereoBars) != numBars {
		v.stereoBars = make([]float32, numBars)
		v.stereoPeaks = make([]float32, numBars)
		v.stereoHold = make([]int, numBars)
		v.stereoFall = make([]float32, numBars)
	}

	if isActive {
		trackBPS := v.CurrentProfile.BPM / 60.0
		beatInterval := float64(1.0 / math.Max(0.1, float64(trackBPS)))
		halfPoint := numBars / 2

		kickPhase := math.Mod(pos, beatInterval) / beatInterval
		kick := float32(math.Exp(-kickPhase*(5.5+float64(v.CurrentProfile.RhythmSwing)*4.0))) * v.CurrentProfile.BassWeight

		snarePhase := math.Mod(pos+beatInterval*0.5, beatInterval) / beatInterval
		snare := float32(math.Exp(-snarePhase*8.0)) * v.CurrentProfile.MidWeight

		hihatPhase := math.Mod(pos, beatInterval*0.25) / (beatInterval * 0.25)
		hihat := float32(math.Exp(-hihatPhase*12.0)) * v.CurrentProfile.TrebleWeight

		for i := 0; i < numBars; i++ {
			div := numBars - 1
			if div < 1 {
				div = 1
			}
			norm := float32(i) / float32(div)
			channelEnergy := liveLeft
			if i >= halfPoint {
				channelEnergy = liveRight
			}
			if !stats.Valid {
				channelEnergy = liveRMS
			}

			tilt := (0.95 * v.CurrentProfile.BassWeight) - (norm * 0.38 * (2.0 - v.CurrentProfile.TrebleWeight))

			w1 := float32(math.Sin(pos*(float64(trackBPS)*3.2*float64(livePitch)) + float64(i)*0.28))
			w2 := float32(math.Cos(pos*(float64(trackBPS)*5.8*float64(livePitch)) - float64(i)*0.52))
			w3 := float32(math.Sin(pos*(float64(trackBPS)*1.4) + float64(i)*0.12))
			wave := (w1*0.42 + w2*0.36 + w3*0.22)

			bassBoost := kick * float32(math.Max(0.0, float64(1.0-norm*2.8))) * (0.4 + livePeak*0.7)
			midBoost := snare * float32(math.Max(0.0, float64(1.0-float32(math.Abs(float64(norm-0.48)))*3.0))) * (0.3 + livePeak*0.6)
			trebleSparkle := (hihat*0.30 + float32(rand.Intn(10))*0.01) * float32(math.Max(0.0, float64(norm-0.50))) * 2.0 * (0.3 + livePeak*0.7)

			rawEnergy := (tilt*(0.28+0.44*wave)*channelEnergy*1.6 + bassBoost*0.70 + midBoost*0.50 + trebleSparkle) * vol
			if stats.Valid {
				rawEnergy *= (0.25 + livePeak*0.85)
			}
			if rawEnergy < 0.02 {
				rawEnergy = 0.02
			}
			if rawEnergy > 1.0 {
				rawEnergy = 1.0
			}

			target := rawEnergy * maxSubLevels
			if target > v.stereoBars[i] {
				v.stereoBars[i] += (target - v.stereoBars[i]) * attackSpeed
			} else {
				v.stereoBars[i] -= (v.stereoBars[i] - target) * decaySpeed
			}

			if v.stereoBars[i] >= v.stereoPeaks[i] {
				v.stereoPeaks[i] = v.stereoBars[i]
				holdCount := 5
				if stats.Valid && livePeak < 0.3 {
					holdCount = 2
				}
				v.stereoHold[i] = holdCount
				v.stereoFall[i] = 0.0
			} else {
				if v.stereoHold[i] > 0 {
					v.stereoHold[i]--
				} else {
					step := float32(0.55)
					if stats.Valid && livePeak < 0.3 {
						step = 0.35
					}
					v.stereoFall[i] += step
					v.stereoPeaks[i] -= v.stereoFall[i]
					if v.stereoPeaks[i] < v.stereoBars[i] {
						v.stereoPeaks[i] = v.stereoBars[i]
					}
				}
			}
		}
	} else {
		for i := 0; i < numBars; i++ {
			v.stereoBars[i] *= 0.85
			v.stereoPeaks[i] *= 0.85
			if v.stereoBars[i] < 0.5 {
				v.stereoBars[i] = 0.0
			}
			if v.stereoPeaks[i] < 0.5 {
				v.stereoPeaks[i] = 0.0
			}
		}
	}

	// Grid generation: drawH rows, drawW columns (reusable buffer)
	grid := v.PrepareGrid(drawW, drawH)

	fullBase := v.cachedBlocks.FullBase
	fullMid := v.cachedBlocks.FullMid
	fullHigh := v.cachedBlocks.FullHigh
	peakChar := v.cachedBlocks.Peak

	for i := 0; i < numBars; i++ {
		val := int(v.stereoBars[i])
		fullCells := val / 8
		rem := val % 8
		peakCell := int(v.stereoPeaks[i]) / 8
		barX := startX + i*slotW

		for y := 0; y < drawH; y++ {
			drawY := drawH - 1 - y
			tier := 0
			if y >= drawH*2/3 {
				tier = 2
			} else if y >= drawH/3 {
				tier = 1
			}

			if y < fullCells {
				char := fullBase
				if tier == 1 {
					char = fullMid
				} else if tier == 2 {
					char = fullHigh
				}
				for k := 0; k < barW && (barX+k) < drawW; k++ {
					grid[drawY][barX+k] = char
				}
			} else if y == fullCells && rem > 0 {
				char := v.cachedBlocks.Base[rem]
				if tier == 1 {
					char = v.cachedBlocks.Mid[rem]
				} else if tier == 2 {
					char = v.cachedBlocks.High[rem]
				}
				for k := 0; k < barW && (barX+k) < drawW; k++ {
					grid[drawY][barX+k] = char
				}
			} else if y == peakCell && peakCell > fullCells && peakCell < drawH {
				for k := 0; k < barW && (barX+k) < drawW; k++ {
					grid[drawY][barX+k] = peakChar
				}
			}
		}
	}

	return v.BuildLines(drawH)
}
