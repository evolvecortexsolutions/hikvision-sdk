package client

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/evolvecortexsolutions/hikvision-go-wrapper/internal/bridge"
)

type Config struct {
	IP       string
	Port     uint16
	Username string
	Password string
}

type Client struct {
	mu       sync.Mutex
	cfg      Config
	userID   int32
	loggedIn bool
}

// New creates a new Client instance.
func New(cfg Config) *Client {
	return &Client{cfg: cfg}
}

// Login logs into the device using NET_DVR_Login_V30.
func (c *Client) Login() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loggedIn {
		return nil
	}

	if err := bridge.InitSDK(); err != nil {
		return fmt.Errorf("bridge init failed: %w", err)
	}

	userID, err := bridge.LoginV30(c.cfg.IP, c.cfg.Port, c.cfg.Username, c.cfg.Password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	c.userID = userID
	c.loggedIn = true
	return nil
}

// Logout logs out the current session.
func (c *Client) Logout() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.loggedIn {
		return nil
	}

	if err := bridge.Logout(c.userID); err != nil {
		return err
	}

	c.loggedIn = false
	c.userID = -1
	return nil
}

// SessionID returns the current session id, or -1 if not logged in.
func (c *Client) SessionID() int32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.loggedIn {
		return -1
	}
	return c.userID
}

// StartTalk begins two-way audio. Returns voice handle.
func (c *Client) StartTalk() (int32, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.loggedIn {
		return -1, errors.New("not logged in")
	}
	return bridge.StartTalk(c.userID)
}

// StopTalk closes a voice channel.
func (c *Client) StopTalk(voiceHandle int32) error {
	return bridge.StopTalk(voiceHandle)
}

// StartPlayback starts a VOD playback session.
func (c *Client) StartPlayback(start, end time.Time, streamType uint8, fileIndex uint32) (int32, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.loggedIn {
		return -1, errors.New("not logged in")
	}
	return bridge.PlayBackByTime(c.userID, start, end, streamType, fileIndex)
}

// StopPlayback stops a VOD playback session.
func (c *Client) StopPlayback(playHandle int32) error {
	return bridge.StopPlayback(playHandle)
}

// Cleanup relays SDK cleanup.
func Cleanup() {
	bridge.CleanupSDK()
}
