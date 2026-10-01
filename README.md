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
- **Real-Time Audio Visualizer**: Fluid Cava wave spectrum with Monstercat neighbor smoothing and multi-tier theme gradients.
- **Curated Color Themes**: `Midnight` (default), `Nord`, `Matrix`, `HyDE`, `Gruvbox`, and `Slate`.
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
| <kbd>T</kbd> | Switch theme (`Midnight` &rarr; `Nord` &rarr; `Matrix` &rarr; `HyDE` &rarr; `Gruvbox` &rarr; `Slate`) |
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

### Quick Install (Recommended)

Install the pre-built binary directly for your platform (Linux / macOS):

```bash
curl -fsSL https://raw.githubusercontent.com/Swadesh-c0de/vibe-fi-go/main/install.sh | bash
```

To install system-wide (`/usr/local/bin`):
```bash
curl -fsSL https://raw.githubusercontent.com/Swadesh-c0de/vibe-fi-go/main/install.sh | bash -s -- --global
```

---

### Build from Source

```bash
# Clone the repository
git clone https://github.com/Swadesh-c0de/vibe-fi-go.git
cd vibe-fi-go

# Run automated installer
./install.sh

# Or build manually using Make
make build
make install
```
