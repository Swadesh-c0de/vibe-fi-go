# Contributing to Vibe-Fi v2

Thanks for checking out the source. Vibe-Fi v2 is the rewritten Go edition (transitioned from the original C++ player), built on Charm's [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the terminal UI and CGO bindings to **`libmpv`** for audio playback.

Whether you're fixing a bug, adding a new theme, or improving audio performance, this guide covers the local setup, codebase layout, and pull request workflow.

---

### Prerequisites & Toolchain

To build Vibe-Fi from source, you need **Go 1.22+**, **pkg-config**, and **libmpv** development libraries.

- **Arch Linux / Manjaro:**
  ```bash
  sudo pacman -S mpv pkg-config yt-dlp
  ```
- **Debian / Ubuntu / Pop!_OS:**
  ```bash
  sudo apt install libmpv-dev pkg-config yt-dlp
  ```
- **Fedora:**
  ```bash
  sudo dnf install mpv-libs-devel pkgconfig yt-dlp
  ```
- **macOS:**
  ```bash
  brew install mpv pkg-config yt-dlp
  ```

> **Note on CGO:** Because Vibe-Fi binds directly to `libmpv`, `CGO_ENABLED=1` is required. Ensure your C compiler (`gcc` or `clang`) and `pkg-config` can locate `mpv.pc`. On macOS, Homebrew paths are exported automatically by the Makefile.

---

### Local Development Workflow

Clone your fork and navigate into the repository:

```bash
git clone https://github.com/<your-username>/vibe-fi-go.git
cd vibe-fi-go
```

Use the Makefile targets during development:

```bash
# Compile binary to build/vibe
make build

# Build and launch immediately
make run

# Run code verification
go vet ./...

# Clean build artifacts
make clean
```

---

### Codebase Architecture

The project separates audio playback, business services, and the Bubble Tea TUI into distinct packages:

```text
cmd/vibe/
└── main.go                     # Entry point, flag parsing, player initialization

internal/
├── config/
│   ├── config.go               # Base directory paths & user settings
│   └── state.go                # Session persistence (~/.vibe-fi/state.ini)
├── eventbus/
│   └── eventbus.go             # Pub/Sub event bus for cross-subsystem messaging
├── integration/
│   ├── discord/
│   │   └── discord.go          # Discord Rich Presence IPC socket client
│   └── mpris/
│       └── mpris.go            # Linux D-Bus MPRIS media keys & playerctl integration
├── player/
│   ├── player.go               # AudioPlayer interface & playback event definitions
│   ├── mpv.go                  # libmpv CGO wrapper & live acoustic stats (RMS/Peak)
│   └── mock.go                 # Mock player implementation for headless testing
├── service/
│   ├── library/                # Recursive filesystem audio scanner & metadata loader
│   ├── lyrics/                 # lrclib.net fetcher, companion .lrc parser & disk cache
│   ├── playlist/               # Plain-text playlist manager (~/.vibe-fi/playlists/)
│   ├── search/                 # yt-dlp search engine & in-memory stream URL cache
│   └── updater/                # Background GitHub release checker & binary updater
├── tui/
│   ├── app.go                  # Root Bubble Tea model & event loop
│   ├── messages.go             # Bubble Tea TEA message definitions & async commands
│   ├── navigation_controller.go# View routing, search results, and confirmation modals
│   ├── playback_controller.go  # Playback controls, seek, volume, and queue actions
│   ├── components/             # Reusable UI widgets (border, helpbar, modal, statusbar)
│   ├── theme/                  # Theme definitions, Lip Gloss styles, themes.json loader
│   ├── views/                  # Stateless view renderers (Playback, Library, Queue, etc.)
│   └── visualizer/             # Cava spectrum math, Monstercat smoothing & block rendering
└── utils/
    ├── bottle/                 # Isolated runtime dependency installer (yt-dlp)
    ├── mem/                    # Platform process RSS memory stats (Linux & macOS)
    ├── net/                    # Network connectivity checks & HTTP clients
    └── stringutil/             # Unicode rune width & terminal text layout helpers
```

#### How State Flows
1. **Audio Engine (`internal/player/`)**: `MPVPlayer` manages the `libmpv` instance via CGO and polls playback events. An `astats` audio filter continuously computes real-time RMS, peak levels, and zero-crossing rates for the visualizer.
2. **Event Bus (`internal/eventbus/`)**: A decoupled pub/sub bus distributes audio events (state changes, progress ticks, track finishes) across the TUI, Discord RPC, and MPRIS without tight coupling.
3. **Controllers (`internal/tui/`)**: User input received in `app.go` delegates to `navigation_controller.go` for view transitions/dialogs and `playback_controller.go` for player interaction, queue management, and lyric fetching.
4. **Views (`internal/tui/views/`)**: Views are stateless rendering functions that take immutable view state structs and return styled strings built with Lip Gloss.

---

### Submitting Changes

1. **Keep Pull Requests Focused**: A PR should address one specific feature or bug fix. Avoid mixing unrelated refactors into feature PRs.
2. **Verify Before Pushing**:
   ```bash
   go vet ./...
   make build
   ```
   Ensure the code compiles without warnings or errors.
3. **Commit Messages**: Use concise, conventional commit prefixes:
   - `feat:` for new capabilities
   - `fix:` for bug fixes
   - `refactor:` for code restructuring without behavior changes
   - `perf:` for performance optimizations
4. **Open the PR**: Push your branch and open a pull request against `main`. Provide a brief summary of what changed and what issue it addresses.

---

### Contributing Themes

Vibe-Fi includes built-in themes in `internal/tui/theme/theme.go`. If you're designing a palette for the project:
- Focus on low-contrast, muted tones that are gentle on the eyes in dark environments.
- Avoid oversaturated, bright neon colors.
- Ensure the three visualizer gradient tiers (`VizBaseColor`, `VizMidColor`, `VizHighColor`) blend harmoniously.
- Test your palette across all views: Home, Playback, Queue, Library, and Modals.

---

### Questions or Issues?

Found a bug or have a suggestion? Open an issue on GitHub. Pull requests and constructive feedback are always appreciated.
