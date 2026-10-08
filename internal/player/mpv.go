package player

/*
#cgo pkg-config: mpv
#include <mpv/client.h>
#include <stdlib.h>
#include <string.h>
#include <locale.h>

typedef struct {
    float rms_overall;
    float peak_overall;
    float rms_left;
    float rms_right;
    float zero_crossings;
    int valid;
} c_astats_t;

static inline float c_db_to_linear(float db) {
    if (db <= -55.0f) return 0.0f;
    if (db >= 0.0f) return 1.0f;
    float norm = (db + 50.0f) / 50.0f;
    if (norm < 0.0f) norm = 0.0f;
    if (norm > 1.0f) norm = 1.0f;
    return norm;
}

static inline c_astats_t extract_astats(mpv_handle* mpv) {
    c_astats_t res = {0.0f, 0.0f, 0.0f, 0.0f, 0.05f, 0};
    if (!mpv) return res;

    mpv_node node;
    if (mpv_get_property(mpv, "af-metadata/astats", MPV_FORMAT_NODE, &node) >= 0) {
        if (node.format == MPV_FORMAT_NODE_MAP && node.u.list) {
            res.valid = 1;
            float raw_rms_overall = -60.0f;
            float raw_peak_overall = -60.0f;
            float raw_rms_left = -60.0f;
            float raw_rms_right = -60.0f;

            for (int i = 0; i < node.u.list->num; ++i) {
                const char* key = node.u.list->keys[i];
                if (node.u.list->values[i].format != MPV_FORMAT_STRING) continue;
                const char* val_str = node.u.list->values[i].u.string;
                if (!val_str) continue;

                if (strcmp(key, "lavfi.astats.Overall.RMS_level") == 0) {
                    raw_rms_overall = (float)atof(val_str);
                } else if (strcmp(key, "lavfi.astats.Overall.Peak_level") == 0) {
                    raw_peak_overall = (float)atof(val_str);
                } else if (strcmp(key, "lavfi.astats.1.RMS_level") == 0) {
                    raw_rms_left = (float)atof(val_str);
                } else if (strcmp(key, "lavfi.astats.2.RMS_level") == 0) {
                    raw_rms_right = (float)atof(val_str);
                } else if (strstr(key, "Zero_crossings_rate") != NULL) {
                    res.zero_crossings = (float)atof(val_str);
                }
            }

            res.rms_overall = c_db_to_linear(raw_rms_overall);
            res.peak_overall = c_db_to_linear(raw_peak_overall);
            res.rms_left = c_db_to_linear(raw_rms_left);
            res.rms_right = c_db_to_linear(raw_rms_right);

            if (res.rms_right <= 0.001f) {
                res.rms_right = (res.rms_left > 0.001f) ? res.rms_left : res.rms_overall;
            }
        }
        mpv_free_node_contents(&node);
    }
    return res;
}

static inline int mpv_cmd_loadfile(mpv_handle* mpv, const char* path, const char* mode) {
    const char* cmd[] = {"loadfile", path, mode, NULL};
    return mpv_command(mpv, cmd);
}

static inline int mpv_cmd_seek(mpv_handle* mpv, const char* sec_str) {
    const char* cmd[] = {"seek", sec_str, "relative", NULL};
    return mpv_command(mpv, cmd);
}

static inline int mpv_cmd_stop(mpv_handle* mpv) {
    const char* cmd[] = {"stop", NULL};
    return mpv_command(mpv, cmd);
}

static inline int get_mpv_end_file_reason(mpv_event* ev) {
    if (!ev || !ev->data) return -1;
    mpv_event_end_file* eef = (mpv_event_end_file*)ev->data;
    return eef->reason;
}

static inline const char* get_mpv_end_file_error(mpv_event* ev) {
    if (!ev || !ev->data) return "";
    mpv_event_end_file* eef = (mpv_event_end_file*)ev->data;
    return mpv_error_string(eef->error);
}
*/
import "C"

import (
	"fmt"
	"strconv"
	"sync"
	"unsafe"
	"vibe-fi/internal/utils/bottle"
	"vibe-fi/internal/utils/mem"
)

var (
	cPropTimePos  = C.CString("time-pos")
	cPropDuration = C.CString("duration")
	cPropPause    = C.CString("pause")
	cPropIdle     = C.CString("idle-active")
	cPropCache    = C.CString("paused-for-cache")
	cPropVolume   = C.CString("volume")
)

// MPVPlayer implements AudioPlayer wrapping libmpv via CGO.
type MPVPlayer struct {
	mu            sync.Mutex
	mpv           *C.mpv_handle
	trackFinished bool
	playbackError bool
	loadingActive bool
	lastError     string
}

// NewMPVPlayer creates and configures a new libmpv instance.
func NewMPVPlayer() (*MPVPlayer, error) {
	// Invariant 1: Force decimal locale to C for reliable timestamp parsing
	cLocale := C.CString("C")
	defer C.free(unsafe.Pointer(cLocale))
	C.setlocale(C.LC_NUMERIC, cLocale)

	mpv := C.mpv_create()
	if mpv == nil {
		return nil, fmt.Errorf("failed to create libmpv context")
	}

	p := &MPVPlayer{mpv: mpv}

	// Audio-only & performance flags
	p.setOption("config", "no")
	p.setOption("vo", "null")
	p.setOption("vd", "null")
	p.setOption("sub", "no")
	p.setOption("sub-auto", "no")
	p.setOption("embeddedfonts", "no")
	p.setOption("audio-display", "no")
	p.setOption("ytdl", "yes")
	p.setOption("ytdl-format", "251/140/bestaudio[ext=m4a]/bestaudio[ext=webm]/bestaudio/best")

	p.setOption("video", "no")
	p.setOption("osc", "no")
	p.setOption("load-stats-overlay", "no")
	p.setOption("load-console", "no")
	p.setOption("load-osd-console", "no")
	p.setOption("load-context-menu", "no")
	p.setOption("load-positioning", "no")
	p.setOption("load-select", "no")
	p.setOption("load-commands", "no")
	p.setOption("load-auto-profiles", "no")

	// Audio output fallback (Linux: pipewire/pulse/alsa, macOS: coreaudio, Windows: wasapi)
	p.setOption("ao", "pipewire,pulse,alsa,coreaudio,wasapi,audiotrack,")

	// Network resilience & buffer optimization
	p.setOption("stream-lavf-o", "reconnect=1,reconnect_delay_max=5")
	p.setOption("network-timeout", "30")
	p.setOption("demuxer-max-bytes", "1024KiB")
	p.setOption("demuxer-max-back-bytes", "128KiB")
	p.setOption("demuxer-readahead-secs", "5")

	// Locate yt-dlp & configure anti-bot extractor options
	p.setOption("ytdl-raw-options", "extractor-args=youtube:player_client=android")
	if ytdlPath := bottle.FindExecutable("yt-dlp"); ytdlPath != "" {
		p.setOption("script-opts", "ytdl_hook-ytdl_path="+ytdlPath)
	}

	// Attach real-time audio statistics filter
	p.setOption("af", "@astats:lavfi=[astats=metadata=1:reset=1:length=0.04]")

	if status := C.mpv_initialize(mpv); status < 0 {
		C.mpv_terminate_destroy(mpv)
		return nil, fmt.Errorf("failed to initialize mpv: %s", C.GoString(C.mpv_error_string(status)))
	}

	mem.TrimMemory()

	return p, nil
}

func (p *MPVPlayer) setOption(key, val string) {
	if p.mpv == nil {
		return
	}
	cKey := C.CString(key)
	cVal := C.CString(val)
	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cVal))
	C.mpv_set_option_string(p.mpv, cKey, cVal)
}

func (p *MPVPlayer) Load(path string, mode string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.mpv == nil {
		return fmt.Errorf("player closed")
	}

	mem.PeriodicTrim()

	p.trackFinished = false
	p.playbackError = false
	p.loadingActive = true
	p.lastError = ""

	cPath := C.CString(path)
	cMode := C.CString(mode)
	defer C.free(unsafe.Pointer(cPath))
	defer C.free(unsafe.Pointer(cMode))

	ret := C.mpv_cmd_loadfile(p.mpv, cPath, cMode)
	if ret < 0 {
		return fmt.Errorf("mpv loadfile error: %s", C.GoString(C.mpv_error_string(ret)))
	}
	return nil
}

func (p *MPVPlayer) Play() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil {
		return nil
	}
	flag := C.int(0)
	ret := C.mpv_set_property(p.mpv, cPropPause, C.MPV_FORMAT_FLAG, unsafe.Pointer(&flag))
	if ret < 0 {
		return fmt.Errorf("mpv play error: %s", C.GoString(C.mpv_error_string(ret)))
	}
	return nil
}

func (p *MPVPlayer) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil {
		return nil
	}
	flag := C.int(1)
	ret := C.mpv_set_property(p.mpv, cPropPause, C.MPV_FORMAT_FLAG, unsafe.Pointer(&flag))
	if ret < 0 {
		return fmt.Errorf("mpv pause error: %s", C.GoString(C.mpv_error_string(ret)))
	}
	return nil
}

func (p *MPVPlayer) TogglePause() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil {
		return nil
	}
	var flag C.int
	if C.mpv_get_property(p.mpv, cPropPause, C.MPV_FORMAT_FLAG, unsafe.Pointer(&flag)) >= 0 {
		if flag == 0 {
			flag = 1
		} else {
			flag = 0
		}
		C.mpv_set_property(p.mpv, cPropPause, C.MPV_FORMAT_FLAG, unsafe.Pointer(&flag))
	}
	return nil
}

func (p *MPVPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.trackFinished = false
	p.playbackError = false
	p.loadingActive = false
	if p.mpv == nil {
		return nil
	}
	ret := C.mpv_cmd_stop(p.mpv)
	if ret < 0 {
		return fmt.Errorf("mpv stop error: %s", C.GoString(C.mpv_error_string(ret)))
	}
	return nil
}

func (p *MPVPlayer) Seek(seconds float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil {
		return nil
	}
	secStr := strconv.FormatFloat(seconds, 'f', 2, 64)
	cSec := C.CString(secStr)
	defer C.free(unsafe.Pointer(cSec))
	ret := C.mpv_cmd_seek(p.mpv, cSec)
	if ret < 0 {
		return fmt.Errorf("mpv seek error: %s", C.GoString(C.mpv_error_string(ret)))
	}
	return nil
}

func (p *MPVPlayer) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil || p.isIdleLocked() || p.isPausedLocked() || p.loadingActive || p.isBufferingLocked() {
		return false
	}
	var pos C.double
	if C.mpv_get_property(p.mpv, cPropTimePos, C.MPV_FORMAT_DOUBLE, unsafe.Pointer(&pos)) < 0 {
		return false
	}
	return true
}

func (p *MPVPlayer) IsPaused() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isPausedLocked()
}

func (p *MPVPlayer) isPausedLocked() bool {
	if p.mpv == nil || p.isIdleLocked() {
		return false
	}
	var flag C.int
	if C.mpv_get_property(p.mpv, cPropPause, C.MPV_FORMAT_FLAG, unsafe.Pointer(&flag)) < 0 {
		return false
	}
	return flag != 0
}

func (p *MPVPlayer) IsIdle() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isIdleLocked()
}

func (p *MPVPlayer) isIdleLocked() bool {
	if p.mpv == nil {
		return true
	}
	var flag C.int = 1
	if C.mpv_get_property(p.mpv, cPropIdle, C.MPV_FORMAT_FLAG, unsafe.Pointer(&flag)) < 0 {
		return true
	}
	return flag != 0
}

func (p *MPVPlayer) IsLoading() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil {
		return false
	}
	if p.loadingActive {
		return true
	}
	if !p.isIdleLocked() && !p.isPausedLocked() {
		var pos C.double
		if C.mpv_get_property(p.mpv, cPropTimePos, C.MPV_FORMAT_DOUBLE, unsafe.Pointer(&pos)) < 0 {
			return true
		}
	}
	return false
}

func (p *MPVPlayer) IsBuffering() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isBufferingLocked()
}

func (p *MPVPlayer) isBufferingLocked() bool {
	if p.mpv == nil {
		return false
	}
	var flag C.int
	if C.mpv_get_property(p.mpv, cPropCache, C.MPV_FORMAT_FLAG, unsafe.Pointer(&flag)) >= 0 && flag != 0 {
		return true
	}
	return false
}

func (p *MPVPlayer) Position() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil {
		return 0
	}
	var pos C.double
	if C.mpv_get_property(p.mpv, cPropTimePos, C.MPV_FORMAT_DOUBLE, unsafe.Pointer(&pos)) < 0 {
		return 0
	}
	return float64(pos)
}

func (p *MPVPlayer) Duration() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil {
		return 0
	}
	var dur C.double
	if C.mpv_get_property(p.mpv, cPropDuration, C.MPV_FORMAT_DOUBLE, unsafe.Pointer(&dur)) < 0 {
		return 0
	}
	return float64(dur)
}

func (p *MPVPlayer) Volume() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil {
		return 100
	}
	var vol C.double = 100.0
	if C.mpv_get_property(p.mpv, cPropVolume, C.MPV_FORMAT_DOUBLE, unsafe.Pointer(&vol)) < 0 {
		return 100
	}
	return int(vol)
}

func (p *MPVPlayer) SetVolume(volume int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil {
		return nil
	}
	if volume < 0 {
		volume = 0
	}
	if volume > 150 {
		volume = 150
	}
	vol := C.double(volume)
	C.mpv_set_property(p.mpv, cPropVolume, C.MPV_FORMAT_DOUBLE, unsafe.Pointer(&vol))
	return nil
}

func (p *MPVPlayer) GetMetadata(key string) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil {
		return ""
	}
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	val := C.mpv_get_property_string(p.mpv, cKey)
	if val != nil {
		res := C.GoString(val)
		C.mpv_free(unsafe.Pointer(val))
		return res
	}
	return ""
}

func (p *MPVPlayer) SetProperty(name, value string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil {
		return nil
	}
	cName := C.CString(name)
	cVal := C.CString(value)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cVal))
	ret := C.mpv_set_property_string(p.mpv, cName, cVal)
	if ret < 0 {
		return fmt.Errorf("mpv set_property error: %s", C.GoString(C.mpv_error_string(ret)))
	}
	return nil
}

func (p *MPVPlayer) GetAudioStats() AudioLevelStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv == nil {
		return AudioLevelStats{}
	}
	cStats := C.extract_astats(p.mpv)
	return AudioLevelStats{
		RMSOverall:    float32(cStats.rms_overall),
		PeakOverall:   float32(cStats.peak_overall),
		RMSLeft:       float32(cStats.rms_left),
		RMSRight:      float32(cStats.rms_right),
		ZeroCrossings: float32(cStats.zero_crossings),
		Valid:         cStats.valid != 0,
	}
}

func (p *MPVPlayer) PollEvents() []PlayerEvent {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.mpv == nil {
		return nil
	}

	var events []PlayerEvent
	for {
		if p.mpv == nil {
			break
		}
		ev := C.mpv_wait_event(p.mpv, 0)
		if ev == nil || ev.event_id == C.MPV_EVENT_NONE {
			break
		}

		switch ev.event_id {
		case C.MPV_EVENT_START_FILE:
			p.loadingActive = true
			p.trackFinished = false
			p.playbackError = false
			p.lastError = ""
			events = append(events, PlayerEvent{Type: EventStartFile})

		case C.MPV_EVENT_FILE_LOADED:
			p.loadingActive = false
			events = append(events, PlayerEvent{Type: EventFileLoaded})

		case C.MPV_EVENT_END_FILE:
			p.loadingActive = false
			reason := C.get_mpv_end_file_reason(ev)
			errStr := C.GoString(C.get_mpv_end_file_error(ev))

			// Invariant 6: Distinguish EOF from ERROR to prevent runaway autoplay skipping
			switch reason {
			case C.MPV_END_FILE_REASON_EOF:
				p.trackFinished = true
				events = append(events, PlayerEvent{Type: EventEndFileEOF})
			case C.MPV_END_FILE_REASON_ERROR:
				p.playbackError = true
				p.lastError = errStr
				events = append(events, PlayerEvent{Type: EventEndFileError, Error: errStr})
			case C.MPV_END_FILE_REASON_STOP:
				events = append(events, PlayerEvent{Type: EventEndFileStop})
			}
		}
	}

	return events
}

func (p *MPVPlayer) HasTrackFinished() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.trackFinished
}

func (p *MPVPlayer) ConsumeTrackFinished() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.trackFinished {
		p.trackFinished = false
		return true
	}
	return false
}

func (p *MPVPlayer) HasPlaybackError() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playbackError
}

func (p *MPVPlayer) ConsumePlaybackError() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.playbackError {
		p.playbackError = false
		return true
	}
	return false
}

func (p *MPVPlayer) GetLastError() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastError
}

func (p *MPVPlayer) ClearPlaybackFlags() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.trackFinished = false
	p.playbackError = false
	p.loadingActive = false
	p.lastError = ""
}

func (p *MPVPlayer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mpv != nil {
		handle := p.mpv
		p.mpv = nil
		C.mpv_terminate_destroy(handle)
	}
	return nil
}
