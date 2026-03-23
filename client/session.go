//go:build linux

package client

import (
	"fmt"
	"sync"
	"time"
)

// SessionState tracks the lifecycle of a device session.
type SessionState int

const (
	SessionStateConnected SessionState = iota
	SessionStateTalking
	SessionStateStopped
)

// String returns the string representation of SessionState.
func (s SessionState) String() string {
	switch s {
	case SessionStateConnected:
		return "connected"
	case SessionStateTalking:
		return "talking"
	case SessionStateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// AudioCallback is a function that receives audio data from the device.
type AudioCallback func(handle int32, data []byte) error

// DeviceSession manages a single device connection.
type DeviceSession struct {
	mu               sync.RWMutex
	config           Config
	client           *Client
	state            SessionState
	talkHandle       int32
	audioCallback    AudioCallback
	audioStopCh      chan struct{}
	audioWg          sync.WaitGroup
	lastActivityTime time.Time
	createdAt        time.Time
	talkRestartCount int
}

// NewDeviceSession creates a new device session.
func NewDeviceSession(cfg Config) *DeviceSession {
	return &DeviceSession{
		config:           cfg,
		state:            SessionStateStopped,
		talkHandle:       -1,
		createdAt:        time.Now(),
		lastActivityTime: time.Now(),
	}
}

// Connect initializes the device connection.
func (ds *DeviceSession) Connect() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.state != SessionStateStopped {
		return fmt.Errorf("session already in state: %s", ds.state)
	}

	cli := New(ds.config)
	if err := cli.Login(); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	ds.client = cli
	ds.state = SessionStateConnected
	ds.lastActivityTime = time.Now()

	// Start keepalive in background
	if err := cli.StartKeepAlive(20 * time.Second); err != nil {
		cli.Logout()
		return fmt.Errorf("keepalive failed: %w", err)
	}

	return nil
}

// Disconnect closes the device connection.
func (ds *DeviceSession) Disconnect() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.state == SessionStateStopped {
		return nil
	}

	// Stop talk if active
	if ds.state == SessionStateTalking {
		_ = ds.stopTalkLocked()
	}

	if ds.client != nil {
		ds.client.StopKeepAlive()
		if err := ds.client.Logout(); err != nil {
			return fmt.Errorf("logout failed: %w", err)
		}
	}

	ds.state = SessionStateStopped
	ds.client = nil
	ds.talkHandle = -1
	ds.lastActivityTime = time.Now()

	return nil
}

// SetAudioCallback sets the callback for receiving audio data.
func (ds *DeviceSession) SetAudioCallback(cb AudioCallback) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.state == SessionStateStopped {
		return fmt.Errorf("session not connected")
	}

	ds.audioCallback = cb
	return nil
}

// StartTalk begins a two-way audio session.
func (ds *DeviceSession) StartTalk() error {
	ds.mu.Lock()

	if ds.state == SessionStateStopped {
		ds.mu.Unlock()
		return fmt.Errorf("session not connected")
	}

	if ds.state == SessionStateTalking {
		ds.mu.Unlock()
		return fmt.Errorf("talk already active")
	}

	if ds.client == nil {
		ds.mu.Unlock()
		return fmt.Errorf("client not initialized")
	}

	// Start talk via device (two-way, callback-driven)
	handle, err := ds.client.StartTalkWithCallback()
	if err != nil {
		ds.mu.Unlock()
		return fmt.Errorf("start talk failed: %w", err)
	}

	ds.talkHandle = handle
	ds.state = SessionStateTalking
	ds.lastActivityTime = time.Now()
	ds.talkRestartCount = 0
	ds.audioStopCh = make(chan struct{})
	ds.mu.Unlock()

	return nil
}

// StartTalkSendOnly starts microphone-only mode (no playback) and avoids speaker path checks.
func (ds *DeviceSession) StartTalkSendOnly() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.state == SessionStateStopped {
		return fmt.Errorf("session not connected")
	}

	if ds.state == SessionStateTalking {
		return fmt.Errorf("talk already active")
	}

	if ds.client == nil {
		return fmt.Errorf("client not initialized")
	}

	handle, err := ds.client.StartTalk()
	if err != nil {
		return fmt.Errorf("start send-only talk failed: %w", err)
	}

	ds.talkHandle = handle
	ds.state = SessionStateTalking
	ds.lastActivityTime = time.Now()
	ds.talkRestartCount = 0
	return nil
}

// StopTalk closes the audio session.
func (ds *DeviceSession) StopTalk() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	return ds.stopTalkLocked()
}

// stopTalkLocked is the internal version without lock (assumes lock is held).
func (ds *DeviceSession) stopTalkLocked() error {
	if ds.state != SessionStateTalking {
		return nil // idempotent
	}

	if ds.audioStopCh != nil {
		close(ds.audioStopCh)
		ds.audioStopCh = nil
	}

	ds.audioWg.Wait()

	if ds.talkHandle >= 0 {
		if err := ds.client.StopTalk(ds.talkHandle); err != nil {
			return fmt.Errorf("stop talk failed: %w", err)
		}
	}

	ds.talkHandle = -1
	ds.state = SessionStateConnected
	ds.lastActivityTime = time.Now()

	return nil
}

// SendAudio sends audio data to the device.
func (ds *DeviceSession) SendAudio(data []byte) error {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if ds.state != SessionStateTalking {
		return fmt.Errorf("talk not active, session state: %s", ds.state)
	}

	if ds.talkHandle < 0 {
		return fmt.Errorf("invalid talk handle")
	}

	if ds.client == nil {
		return fmt.Errorf("client not initialized")
	}

	return ds.client.SendAudio(ds.talkHandle, data)
}

// RestartTalk gracefully restarts the talk session.
func (ds *DeviceSession) RestartTalk() error {
	ds.mu.Lock()

	if ds.state == SessionStateStopped {
		ds.mu.Unlock()
		return fmt.Errorf("session not connected")
	}

	wasActive := ds.state == SessionStateTalking
	if wasActive {
		// Stop cleanly
		_ = ds.stopTalkLocked()
	}

	ds.talkRestartCount++
	ds.mu.Unlock()

	if wasActive {
		// Brief pause to ensure resources are released
		time.Sleep(50 * time.Millisecond)
		// Restart
		return ds.StartTalk()
	}

	return nil
}

// State returns current session state.
func (ds *DeviceSession) State() SessionState {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.state
}

// GetStats returns session statistics.
func (ds *DeviceSession) GetStats() map[string]interface{} {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return map[string]interface{}{
		"state":              ds.state.String(),
		"talk_handle":        ds.talkHandle,
		"created_at":         ds.createdAt,
		"last_activity":      ds.lastActivityTime,
		"uptime":             time.Since(ds.createdAt),
		"talk_restart_count": ds.talkRestartCount,
	}
}

// SessionManager manages multiple device sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*DeviceSession // key: "userID:deviceIP"
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*DeviceSession),
	}
}

// GetOrCreate gets an existing session or creates a new one.
func (sm *SessionManager) GetOrCreate(userID string, cfg Config) (*DeviceSession, error) {
	key := userID + ":" + cfg.IP

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[key]; exists {
		return session, nil
	}

	session := NewDeviceSession(cfg)
	if err := session.Connect(); err != nil {
		return nil, err
	}

	sm.sessions[key] = session
	return session, nil
}

// Get retrieves an existing session.
func (sm *SessionManager) Get(userID string, deviceIP string) (*DeviceSession, error) {
	key := userID + ":" + deviceIP

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[key]
	if !exists {
		return nil, fmt.Errorf("session not found for user %s device %s", userID, deviceIP)
	}

	return session, nil
}

// CloseSession closes a specific session.
func (sm *SessionManager) CloseSession(userID string, deviceIP string) error {
	key := userID + ":" + deviceIP

	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[key]
	if !exists {
		return nil
	}

	if err := session.Disconnect(); err != nil {
		return err
	}

	delete(sm.sessions, key)
	return nil
}

// CloseAllSessions closes all active sessions.
func (sm *SessionManager) CloseAllSessions() error {
	sm.mu.Lock()
	sessions := make([]*DeviceSession, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	sm.sessions = make(map[string]*DeviceSession)
	sm.mu.Unlock()

	var lastErr error
	for _, session := range sessions {
		if err := session.Disconnect(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// ListSessions returns info about all active sessions.
func (sm *SessionManager) ListSessions() map[string]map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]map[string]interface{})
	for key, session := range sm.sessions {
		result[key] = session.GetStats()
	}

	return result
}
