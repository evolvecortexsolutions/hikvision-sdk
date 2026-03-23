//go:build linux

package client

import (
	"fmt"
	"sync"
	"time"

	"github.com/evolvecortexsolutions/hikvision-sdk/bridge"
)

// WebSocketAudioStream manages bidirectional audio over WebSocket.
type WebSocketAudioStream struct {
	mu             sync.Mutex
	handle         int32
	session        *DeviceSession
	sendChan       chan []byte
	closeChan      chan struct{}
	wg             sync.WaitGroup
	isClosed       bool
	lastSendTime   time.Time
	bytesReceived  uint64
	bytesSent      uint64
	frameCountRecv uint64
	frameCountSend uint64
}

// NewWebSocketAudioStream creates a managed audio stream for a device session.
func NewWebSocketAudioStream(session *DeviceSession, bufferSize int) (*WebSocketAudioStream, error) {
	if session == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}

	if bufferSize <= 0 {
		bufferSize = 100 // default buffer
	}

	stream := &WebSocketAudioStream{
		handle:       -1,
		session:      session,
		sendChan:     make(chan []byte, bufferSize),
		closeChan:    make(chan struct{}),
		lastSendTime: time.Now(),
	}

	return stream, nil
}

// Start begins the audio stream session.
func (ws *WebSocketAudioStream) Start(onReceive func([]byte) error) error {
	ws.mu.Lock()

	if ws.isClosed {
		ws.mu.Unlock()
		return fmt.Errorf("stream already closed")
	}

	if ws.handle >= 0 {
		ws.mu.Unlock()
		return fmt.Errorf("stream already started")
	}

	if onReceive == nil {
		ws.mu.Unlock()
		return fmt.Errorf("onReceive callback cannot be nil")
	}

	// Register handler before starting talk to avoid race
	handler := func(data []byte) error {
		ws.mu.Lock()
		ws.bytesReceived += uint64(len(data))
		ws.frameCountRecv++
		ws.mu.Unlock()

		return onReceive(data)
	}

	// Use WebSocket-enabled talk
	handle, err := bridge.StartTalkWithWebSocket(ws.session.client.UserID(), handler)
	if err != nil {
		ws.mu.Unlock()
		return fmt.Errorf("start talk failed: %w", err)
	}

	ws.handle = handle
	ws.wg.Add(1)
	go ws.sendLoop()

	ws.mu.Unlock()

	return nil
}

// sendLoop handles outgoing audio frames.
func (ws *WebSocketAudioStream) sendLoop() {
	defer ws.wg.Done()

	for {
		select {
		case data := <-ws.sendChan:
			ws.mu.Lock()
			if ws.isClosed || ws.handle < 0 {
				ws.mu.Unlock()
				return
			}
			client := ws.session.client
			handle := ws.handle
			ws.mu.Unlock()

			if err := client.SendAudio(handle, data); err != nil {
				fmt.Printf("send audio error: %v\n", err)
			} else {
				ws.mu.Lock()
				ws.bytesSent += uint64(len(data))
				ws.frameCountSend++
				ws.lastSendTime = time.Now()
				ws.mu.Unlock()
			}

		case <-ws.closeChan:
			return
		}
	}
}

// SendAudio queues audio data for transmission to the device.
func (ws *WebSocketAudioStream) SendAudio(data []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.isClosed {
		return fmt.Errorf("stream closed")
	}

	if ws.handle < 0 {
		return fmt.Errorf("stream not started")
	}

	if len(data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}

	select {
	case ws.sendChan <- data:
		return nil
	case <-ws.closeChan:
		return fmt.Errorf("stream closed while sending")
	default:
		return fmt.Errorf("send buffer full")
	}
}

// Stop closes the audio stream.
func (ws *WebSocketAudioStream) Stop() error {
	ws.mu.Lock()

	if ws.isClosed {
		ws.mu.Unlock()
		return nil
	}

	ws.isClosed = true
	close(ws.closeChan)

	if ws.handle >= 0 {
		handle := ws.handle
		ws.handle = -1
		ws.mu.Unlock()

		bridge.UnregisterAudioCallback(handle)
		if err := ws.session.client.StopTalk(handle); err != nil {
			return fmt.Errorf("stop talk failed: %w", err)
		}
	} else {
		ws.mu.Unlock()
	}

	// Wait for send loop to finish
	ws.wg.Wait()

	// Drain remaining queued frames
	close(ws.sendChan)

	return nil
}

// Stats returns stream statistics.
func (ws *WebSocketAudioStream) Stats() map[string]interface{} {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	return map[string]interface{}{
		"handle":          ws.handle,
		"is_closed":       ws.isClosed,
		"bytes_received":  ws.bytesReceived,
		"bytes_sent":      ws.bytesSent,
		"frames_received": ws.frameCountRecv,
		"frames_sent":     ws.frameCountSend,
		"last_send_time":  ws.lastSendTime,
	}
}

// IsActive checks if the stream is actively running.
func (ws *WebSocketAudioStream) IsActive() bool {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	return !ws.isClosed && ws.handle >= 0
}
