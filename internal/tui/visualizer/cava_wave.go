package visualizer

import (
	"math"
	"strings"
	"vibe-fi/internal/player"
	"vibe-fi/internal/tui/theme"
)

func (v *Visualizer) renderCavaWave(drawW, drawH int, p player.AudioPlayer, pos float64, vol float32, stats player.AudioLevelStats, styles theme.Styles) []string {
	isActive := p.IsPlaying() && !p.IsPaused() && !p.IsIdle() && !p.IsLoading()
	maxSubLevels := float32(drawH * 8)

	barW := 1
	if drawW >= 60 {
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

	if len(v.cavaBars) != numBars {
		v.cavaBars = make([]float32, numBars)
		v.cavaPeaks = make([]float32, numBars)
		v.cavaHold = make([]int, numBars)
		v.cavaFall = make([]float32, numBars)
	}

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
		if livePitch > 1.60 {
			livePitch = 1.60
		}
	}

	if isActive {
		trackBPS := v.CurrentProfile.BPM / 60.0
		beatInterval := float64(1.0 / math.Max(0.1, float64(trackBPS)))

		kickPhase := math.Mod(pos, beatInterval) / beatInterval
		kick := float32(math.Exp(-kickPhase*(5.2+float64(v.CurrentProfile.RhythmSwing)*3.5))) * v.CurrentProfile.BassWeight

		snarePhase := math.Mod(pos+beatInterval*0.5, beatInterval) / beatInterval
		snare := float32(math.Exp(-snarePhase*7.5)) * v.CurrentProfile.MidWeight

		hihatPhase := math.Mod(pos, beatInterval*0.25) / (beatInterval * 0.25)
		hihat := float32(math.Exp(-hihatPhase*11.0)) * v.CurrentProfile.TrebleWeight

		targets := make([]float32, numBars)
		for i := 0; i < numBars; i++ {
			t := float32(0.5)
			if numBars > 1 {
				t = float32(i) / float32(numBars-1)
			}

			bassBand := float32(math.Pow(float64(1.0-t), 1.5)) * (kick*0.42 + liveRMS*0.30)
			midBand := float32(math.Exp(-math.Pow(float64(t-0.45)/0.28, 2.0))) * (snare*0.35 + liveRMS*0.25)
			trebleBand := float32(math.Pow(float64(t), 1.3)) * (hihat*0.28 + livePitch*0.20)

			var chBlend float32
			if t < 0.5 {
				chBlend = liveLeft*(1.0-t*0.5) + liveRight*(t*0.5)
			} else {
				chBlend = liveLeft*((1.0-t)*0.5) + liveRight*(0.5+t*0.5)
			}

			acousticWave := float32(math.Sin(pos*(float64(trackBPS)*3.2*float64(livePitch))+float64(i)*0.32)*0.08 +
				math.Cos(pos*(float64(trackBPS)*5.5*float64(livePitch))-float64(i)*0.24)*0.06)

			rawEnergy := (bassBand*0.75 + midBand*0.60 + trebleBand*0.50 + chBlend*0.25 + acousticWave) * vol * v.CurrentProfile.EnergyVariance
			if stats.Valid {
				rawEnergy *= (0.28 + livePeak*0.72)
			}

			if rawEnergy < 0.02 {
				rawEnergy = 0.02
			}
			if rawEnergy > 0.95 {
				rawEnergy = 0.95
			}
			targets[i] = rawEnergy * maxSubLevels
		}

		// Monstercat bidirectional smoothing filter
		smoothed := make([]float32, numBars)
		copy(smoothed, targets)
		for i := 1; i < numBars; i++ {
			if smoothed[i-1]*0.68 > smoothed[i] {
				smoothed[i] = smoothed[i-1] * 0.68
			}
		}
		for i := numBars - 2; i >= 0; i-- {
			if smoothed[i+1]*0.68 > smoothed[i] {
				smoothed[i] = smoothed[i+1] * 0.68
			}
		}

		attackSpeed := float32(0.55)
		decaySpeed := float32(0.16)
		if stats.Valid {
			attackSpeed = 0.24 + livePeak*0.48
			decaySpeed = 0.08 + liveRMS*0.12
		}

		for i := 0; i < numBars; i++ {
			target := smoothed[i]
			if target > v.cavaBars[i] {
				v.cavaBars[i] += (target - v.cavaBars[i]) * attackSpeed
			} else {
				v.cavaBars[i] -= (v.cavaBars[i] - target) * decaySpeed
			}

			if v.cavaBars[i] >= v.cavaPeaks[i] {
				v.cavaPeaks[i] = v.cavaBars[i]
				holdCount := 4
				if stats.Valid && livePeak < 0.35 {
					holdCount = 2
				}
				v.cavaHold[i] = holdCount
				v.cavaFall[i] = 0.0
			} else {
				if v.cavaHold[i] > 0 {
					v.cavaHold[i]--
				} else {
					step := float32(0.50)
					if stats.Valid && livePeak < 0.35 {
						step = 0.35
					}
					v.cavaFall[i] += step
					v.cavaPeaks[i] -= v.cavaFall[i]
					if v.cavaPeaks[i] < v.cavaBars[i] {
						v.cavaPeaks[i] = v.cavaBars[i]
					}
				}
			}
		}
	} else {
		for i := 0; i < numBars; i++ {
			v.cavaBars[i] *= 0.85
			v.cavaPeaks[i] *= 0.85
			if v.cavaBars[i] < 0.5 {
				v.cavaBars[i] = 0.0
			}
			if v.cavaPeaks[i] < 0.5 {
				v.cavaPeaks[i] = 0.0
			}
		}
	}

	// Grid generation: drawH rows, drawW columns
	grid := make([][]string, drawH)
	for y := 0; y < drawH; y++ {
		grid[y] = make([]string, drawW)
		for x := 0; x < drawW; x++ {
			grid[y][x] = " "
		}
	}

	for i := 0; i < numBars; i++ {
		val := int(v.cavaBars[i])
		fullCells := val / 8
		rem := val % 8
		peakCell := int(v.cavaPeaks[i]) / 8
		barX := startX + i*slotW

		for y := 0; y < drawH; y++ {
			drawY := drawH - 1 - y
			style := styles.VizBase
			if y >= drawH*3/4 {
				style = styles.VizHigh
			} else if y >= drawH*2/5 {
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
				for k := 0; k < barW && (barX+k) < drawW; k++ {
					grid[drawY][barX+k] = styles.VizPeak.Render(PeakChar)
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
