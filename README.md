<div align="center">

<h4>Free, open-source terminal music player for Linux & macOS written in Go.</h4>

<hr>

<picture>
  <img alt="vibe-fi logo" src="assets/logo.svg" width="220" />
</picture>
<br><br>

<p align="center">
  <strong>A fast, clean music player for your terminal.</strong><br>
  <em>Stream YouTube audio, play local music files, watch real-time visualizers, and read synced lyrics — built with Go & Bubble Tea.</em>
</p>

</div>

---

## Features

- **Zero-Configuration YouTube Streaming**: Direct audio streaming via `yt-dlp` without opening a browser.
- **Local Audio Library**: Instant recursive browsing of local music folders with format detection (`.flac`, `.mp3`, `.wav`, `.m4a`, `.ogg`, `.opus`, `.aac`, `.alac`, `.aiff`, `.webm`).
- **Synchronized LRC Lyrics**: Live line-by-line scrolling lyrics from `lrclib.net` with offline disk cache.
- **Real-Time Audio Visualizers**:
  - **Cava Wave**: Fluid wave spectrum with Monstercat neighbor smoothing and multi-tier theme gradients.
  - **Neon Flame**: Mirrored equalizer volcano with beat metronome indicator and floating peak crowns.
  - **Stereo Bars**: Classic graphic equalizer showing separate left and right channels.
- **Curated Color Themes**: `Midnight` (default), `Matrix`, `Nord`, and `HyDE`.
- **Linux Media Keys (MPRIS)**: Native D-Bus integration for hardware media keys and `playerctl`.
- **Discord Rich Presence**: Native Unix domain socket IPC displaying current track and artist.
- **Full Backward Compatibility**: Reads and writes standard `~/.vibe-fi/state.ini` and plain text playlists (`Title|URL|Duration`).

---

## Hotkeys

| Key | Action |
| :---: | :--- |
| <kbd>SPACE</kbd> | Play / Pause toggle |
| <kbd>N</kbd> / <kbd>&gt;</kbd> | Next track in queue |
| <kbd>B</kbd> / <kbd>&lt;</kbd> | Previous track in queue |
| <kbd>&larr;</kbd> / <kbd>&rarr;</kbd> | Seek backward / forward 5s |
| <kbd>+</kbd> / <kbd>-</kbd> | Volume up / down |
| <kbd>S</kbd> | Search YouTube |
| <kbd>L</kbd> | Browse local music library |
| <kbd>P</kbd> | Playlists manager |
| <kbd>C</kbd> | View play queue |
| <kbd>U</kbd> | Paste YouTube URL to play |
| <kbd>V</kbd> | Switch visualizer (`Cava Wave` &rarr; `Neon Flame` &rarr; `Stereo Bars`) |
| <kbd>T</kbd> | Switch theme (`Midnight` &rarr; `Matrix` &rarr; `Nord` &rarr; `HyDE`) |
| <kbd>O</kbd> | Toggle autoplay (`ON` / `OFF`) |
| <kbd>R</kbd> | Replay track (or restore session from home screen) |
| <kbd>&uarr;</kbd> / <kbd>&darr;</kbd> | Scroll synced lyrics manually |
| <kbd>ESC</kbd> / <kbd>Q</kbd> | Exit player (prompts confirmation) |

---

## Building & Installation

### Requirements
- **Go** (1.20+)
- **libmpv** (`libmpv-dev` on Debian/Ubuntu, `mpv` on Arch/Fedora/macOS)
- **yt-dlp** (optional; will be automatically installed in `~/.vibe-fi/bottle/bin/` if not present)

### Quick Start

```bash
# Build binary
make build

# Run directly
make run

# Run tests
make test

# Install to ~/.local/bin/vibe
make install
```
