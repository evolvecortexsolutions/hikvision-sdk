package bridge

import (
	"fmt"
	"sync"
)

// AudioHandler manages audio callbacks and WebSocket streaming.
type AudioHandler struct {
	mu              sync.Mutex
	activeCallbacks map[int32]func([]byte) error // handle → callback
}

// NewAudioHandler creates a new audio handler.
func NewAudioHandler() *AudioHandler {
	return &AudioHandler{
		activeCallbacks: make(map[int32]func([]byte) error),
	}
}

// RegisterCallback registers a callback for a talk handle.
func (ah *AudioHandler) RegisterCallback(handle int32, cb func([]byte) error) error {
	ah.mu.Lock()
	defer ah.mu.Unlock()

	if handle < 0 {
		return fmt.Errorf("invalid handle: %d", handle)
	}

	if cb == nil {
		return fmt.Errorf("callback cannot be nil")
	}

	ah.activeCallbacks[handle] = cb
	return nil
}

// UnregisterCallback unregisters a callback for a talk handle.
func (ah *AudioHandler) UnregisterCallback(handle int32) {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	delete(ah.activeCallbacks, handle)
}

// CallAudio invokes the registered callback for a handle.
func (ah *AudioHandler) CallAudio(handle int32, data []byte) error {
	ah.mu.Lock()
	cb, exists := ah.activeCallbacks[handle]
	ah.mu.Unlock()

	if !exists {
		return fmt.Errorf("no callback registered for handle %d", handle)
	}

	if len(data) == 0 {
		return nil // ignore empty frames
	}

	return cb(data)
}

// HandleCount returns the number of active handles.
func (ah *AudioHandler) HandleCount() int {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	return len(ah.activeCallbacks)
}

var (
	globalAudioHandlerMu sync.Mutex
	globalAudioHandler   *AudioHandler
)

// GetAudioHandler returns the global audio handler.
func GetAudioHandler() *AudioHandler {
	globalAudioHandlerMu.Lock()
	defer globalAudioHandlerMu.Unlock()

	if globalAudioHandler == nil {
		globalAudioHandler = NewAudioHandler()
	}

	return globalAudioHandler
}

// RegisterAudioCallback registers a device audio callback (for WebSocket streaming).
// This replaces the old playAudio/aplay logic.
func RegisterAudioCallback(handle int32, cb func([]byte) error) error {
	if cb == nil {
		return fmt.Errorf("callback cannot be nil")
	}
	return GetAudioHandler().RegisterCallback(handle, cb)
}

// UnregisterAudioCallback removes the audio callback for a handle.
func UnregisterAudioCallback(handle int32) {
	GetAudioHandler().UnregisterCallback(handle)
}
