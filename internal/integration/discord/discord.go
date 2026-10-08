package discord

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const defaultClientID = "1198647007788449902"

// Client manages direct Unix domain socket connection to Discord IPC.
type Client struct {
	mu        sync.Mutex
	conn      net.Conn
	clientID  string
	connected bool
}

// NewClient creates a new Discord RPC client.
func NewClient(clientID string) *Client {
	if clientID == "" {
		clientID = defaultClientID
	}
	c := &Client{clientID: clientID}
	c.connect()
	return c
}

func (c *Client) connect() {
	var sockPaths []string

	if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
		for i := 0; i < 10; i++ {
			sockPaths = append(sockPaths, filepath.Join(xdg, fmt.Sprintf("discord-ipc-%d", i)))
		}
	}
	for i := 0; i < 10; i++ {
		sockPaths = append(sockPaths, filepath.Join(os.TempDir(), fmt.Sprintf("discord-ipc-%d", i)))
	}

	for _, p := range sockPaths {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		conn, err := net.DialTimeout("unix", p, 250*time.Millisecond)
		if err == nil {
			c.conn = conn
			c.connected = true
			c.sendHandshake()
			return
		}
	}
}

func (c *Client) sendHandshake() {
	handshake := map[string]interface{}{
		"v":         1,
		"client_id": c.clientID,
	}
	payload, _ := json.Marshal(handshake)
	c.sendFrame(0, payload)
}

func (c *Client) sendFrame(opcode int, payload []byte) {
	if c.conn == nil {
		return
	}
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, int32(opcode))
	_ = binary.Write(&buf, binary.LittleEndian, int32(len(payload)))
	buf.Write(payload)

	_, err := c.conn.Write(buf.Bytes())
	if err != nil {
		_ = c.conn.Close()
		c.conn = nil
		c.connected = false
	}
}

// UpdatePresence updates active song and artist on Discord.
func (c *Client) UpdatePresence(songTitle, artist string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		c.connect()
	}
	if !c.connected {
		return
	}

	activity := map[string]interface{}{
		"details": songTitle,
		"state":   artist,
		"assets": map[string]string{
			"large_image": "vibe_fi_logo",
			"large_text":  "Vibe-Fi Terminal Music Player",
		},
		"timestamps": map[string]int64{
			"start": time.Now().Unix(),
		},
	}

	payload := map[string]interface{}{
		"cmd": "SET_ACTIVITY",
		"args": map[string]interface{}{
			"pid":      os.Getpid(),
			"activity": activity,
		},
		"nonce": fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	data, err := json.Marshal(payload)
	if err == nil {
		c.sendFrame(1, data)
	}
}

// ClearPresence clears presence state.
func (c *Client) ClearPresence() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.connected {
		return
	}

	payload := map[string]interface{}{
		"cmd": "SET_ACTIVITY",
		"args": map[string]interface{}{
			"pid": os.Getpid(),
		},
		"nonce": fmt.Sprintf("%d", time.Now().UnixNano()),
	}
	data, err := json.Marshal(payload)
	if err == nil {
		c.sendFrame(1, data)
	}
}

// Close disconnects the Unix domain socket.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.sendFrame(2, []byte("{}"))
		_ = c.conn.Close()
		c.conn = nil
		c.connected = false
	}
}
