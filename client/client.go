//go:build linux

package client

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/evolvecortexsolutions/hikvision-sdk/bridge"
)

type DeviceType int

const (
	DeviceTypeUnknown DeviceType = iota
	DeviceTypeNVR
	DeviceTypeDVR
)

type Config struct {
	IP         string
	Port       uint16
	Username   string
	Password   string
	DeviceType DeviceType
}

type Client struct {
	mu            sync.Mutex
	cfg           Config
	userID        int32
	loggedIn      bool
	keepAliveStop chan struct{}
	keepAliveWg   sync.WaitGroup
}

// New creates a new Client instance.
func New(cfg Config) *Client {
	if cfg.DeviceType == DeviceTypeUnknown {
		cfg.DeviceType = DeviceTypeNVR
	}
	return &Client{cfg: cfg, userID: -1}
}

// Login logs into the device using NET_DVR_Login_V40.
func (c *Client) Login() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loggedIn {
		return nil
	}

	if err := bridge.InitSDK(); err != nil {
		return fmt.Errorf("bridge init failed: %w", err)
	}

	userID, err := bridge.LoginV40(c.cfg.IP, c.cfg.Port, c.cfg.Username, c.cfg.Password)
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
	c.StopKeepAlive()
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

// UserID returns the internal user ID for low-level bridge operations.
// Returns -1 if not logged in.
func (c *Client) UserID() int32 {
	return c.SessionID()
}

// KeepAlive sends a one-shot keepalive call to the device.
func (c *Client) KeepAlive() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.loggedIn {
		return errors.New("not logged in")
	}
	return bridge.KeepAlive(c.userID)
}

// StartKeepAlive starts periodic keepalive heartbeats.
func (c *Client) StartKeepAlive(interval time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.loggedIn {
		return errors.New("not logged in")
	}
	if interval <= 0 {
		interval = 30 * time.Second
	}
	if c.keepAliveStop != nil {
		return nil
	}
	c.keepAliveStop = make(chan struct{})
	c.keepAliveWg.Add(1)
	go func() {
		defer c.keepAliveWg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := bridge.KeepAlive(c.userID); err != nil {
					fmt.Printf("keepalive failed for user %d: %v\n", c.userID, err)
				}
			case <-c.keepAliveStop:
				return
			}
		}
	}()
	return nil
}

// StopKeepAlive stops the keepalive heartbeat goroutine.
func (c *Client) StopKeepAlive() {
	c.mu.Lock()
	stop := c.keepAliveStop
	c.keepAliveStop = nil
	c.mu.Unlock()
	if stop != nil {
		close(stop)
		c.keepAliveWg.Wait()
	}
}

// GetDVRConfig reads DVR config into the provided buffer.
func (c *Client) GetDVRConfig(command uint32, channel int32, out []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.loggedIn {
		return 0, errors.New("not logged in")
	}
	return bridge.GetDVRConfig(c.userID, command, channel, out)
}

// GetDVRWorkState returns device work state and health info.
func (c *Client) GetDVRWorkState() (bridge.DVRWorkState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.loggedIn {
		return bridge.DVRWorkState{}, errors.New("not logged in")
	}
	return bridge.GetDVRWorkState(c.userID)
}

// GetNetworkConfig returns network configuration including MAC addresses and IP addresses.
func (c *Client) GetNetworkConfig() ([]bridge.NetworkInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.loggedIn {
		return nil, errors.New("not logged in")
	}
	return bridge.GetNetworkConfig(c.userID)
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

// StartTalkWithCallback begins two-way audio and plays device audio locally.
func (c *Client) StartTalkWithCallback() (int32, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.loggedIn {
		return -1, errors.New("not logged in")
	}
	return bridge.StartTalkWithCallback(c.userID)
}

// StopTalk closes a voice channel.
func (c *Client) StopTalk(voiceHandle int32) error {
	return bridge.StopTalk(voiceHandle)
}

// SendAudio sends raw PCM data into the active voice talk channel.
func (c *Client) SendAudio(voiceHandle int32, payload []byte) error {
	if len(payload) == 0 {
		return errors.New("payload cannot be empty")
	}
	return bridge.SendAudio(voiceHandle, payload)
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
