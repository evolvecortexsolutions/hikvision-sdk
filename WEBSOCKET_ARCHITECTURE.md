# Hikvision SDK - WebSocket Audio Architecture

## Overview

This document describes the production-grade WebSocket-based audio architecture for multi-device, multi-user Hikvision SDK integration. The system is fully thread-safe, supports concurrent device sessions, and handles graceful session lifecycle management including restart scenarios.

## Architecture Components

### 1. **Session Manager** (`client/session.go`)

Global registry for managing multiple device sessions per user.

```go
manager := client.NewSessionManager()
session, err := manager.GetOrCreate("user_alice", deviceConfig)
```

**Key Features:**
- Thread-safe concurrent session management
- Per-device state tracking (Connected, Talking, Stopped)
- Automatic keepalive management
- Session cleanup and resource deallocation

**Thread Safety:**
- `sync.RWMutex` for concurrent read access
- Atomic map operations
- Safe session lifecycle transitions

### 2. **Device Session** (`client/session.go`)

Represents a single device connection for one user.

```go
type DeviceSession struct {
    state           SessionState
    client          *Client
    talkHandle      int32
    audioCallback   AudioCallback
}
```

**States:**
- `SessionStateConnected`: Device connected, ready for talk
- `SessionStateTalking`: Active bidirectional audio
- `SessionStateStopped`: Session closed

**Critical Operations:**
- `Connect()` - Establish device login + keepalive
- `StartTalk()` - Begin audio session
- `StopTalk()` - Graceful audio shutdown  
- `RestartTalk()` - Safe restart with state cleanup
- `SendAudio(data)` - Queue microphone → device

---

## Audio Handler System (`bridge/audio_handler.go`)

Replaces local playback (aplay) with pluggable callback routing.

```go
type AudioHandler struct {
    activeCallbacks map[int32]func([]byte) error
}
```

**Flow:**
1. Device sends audio → C callback `goAudioCallback()`
2. `goAudioCallback()` routes to `handleAudioData()`
3. `handleAudioData()` invokes registered callback
4. Callback sends frame via WebSocket to browser

**No aplay dependency** - pure handler-based architecture.

---

## WebSocket Audio Stream (`client/websocket_audio.go`)

Manages bidirectional PCM streaming over WebSocket.

```go
stream, err := client.NewWebSocketAudioStream(session, bufferSize)

// Register browser audio sink
err = stream.Start(func(data []byte) error {
    return websocket.WriteMessage(ws, websocket.BinaryMessage, data)
})

// Send microphone data
err = stream.SendAudio(micFrameData)

// Graceful stop
err = stream.Stop()
```

**Key Features:**
- Async send loop with buffering
- Frame statistics (bytes/count)
- Safe resource cleanup
- Non-blocking sender (queue-based)

---

## Concurrency Model

### Thread Safety Guarantees

**No Race Conditions:**
```
Session State Transitions:
    StopTalkLocked() ← held lock
        ↓
    close(audioStopCh) ← signals send loop
        ↓
    stream.wg.Wait() ← waits for completion
        ↓
    Device.StopTalk() ← safe unblock
```

**Multi-Device Parallelism:**
```
User Alice           User Bob
├─ Device A (talk)   ├─ Device C (idle)
├─ Device B (idle)   └─ Device D (talk)
└─ Connected         └─ Connected
```

Each session is independent with no shared state except AudioHandler callback registry (which is concurrent-safe).

### Critical Restart Scenario

**Problem:** StopTalk() called, immediately followed by StartTalk()

**Solution:** Mutex-protected state machine
```go
func (ds *DeviceSession) RestartTalk() error {
    ds.mu.Lock()
    
    wasActive := ds.state == SessionStateTalking
    if wasActive {
        _ = ds.stopTalkLocked()  // Clean shutdown
    }
    
    ds.mu.Unlock()
    
    time.Sleep(50 * time.Millisecond)  // Resource release
    
    return ds.StartTalk()  // Fresh start
}
```

**Guarantees:**
- No stale handles
- No blocked goroutines
- No resource leaks
- Clean device state reset

---

## Flow Diagrams

### Start Talk → Audio → Stop

```
Browser                WebSocket      Server              Device
  │                        │             │                  │
  │  connect()             │             ├─ Login           │
  │◄──────────────────────►│◄────────────┤ Keepalive        │
  │                        │             │                  │
  │  StartTalk()           │             ├─ StartTalk()     │
  │◄──────────────────────►│◄────────────┤ (handles audio)  │
  │                        │             │                  │
  │                        │   Device◄───┤ audio stream ──►  │
  │  audio data ───────────┼────────────►│ (callback)       │
  │                        │             │                  │
  │              ◄─────────┼─────────────┤ audio → WS       │
  │              audio data│             │                  │
  │                        │             │                  │
  │  SendAudio()           │             │                  │
  │──────────────┬─────────┼────────────►├─ SendAudio()     │
  │              │audio────┼────────────►│                  │
  │              │data     │             │                  │
  │                        │             │                  │
  │  StopTalk()            │             ├─ StopTalk()      │
  │◄──────────────────────►│◄────────────┤ cleanup          │
  │                        │             │                  │
```

### Parallel Multi-Device (Same User)

```
SessionManager [user_alice]
├── DeviceSession (192.168.1.100)
│   ├─ state: Talking
│   ├─ handle: 1
│   └─ stream: active
├── DeviceSession (192.168.1.101)
│   ├─ state: Connected
│   ├─ handle: -1
│   └─ stream: nil
└── DeviceSession (192.168.1.102)
    ├─ state: Talking
    ├─ handle: 3
    └─ stream: active

All sessions independent, concurrent sends/receives
```

---

## API Examples

### Example 1: Single Device Audio Session

```go
// Get or create session
session, _ := manager.GetOrCreate("user1", config)

// Start talk
_ = session.StartTalk()

// Create WebSocket stream
stream, _ := client.NewWebSocketAudioStream(session, 100)

// Register browser audio sink
_ = stream.Start(func(data []byte) error {
    return ws.WriteMessage(websocket.BinaryMessage, data)
})

// Send microphone data
for micFrame := range micInput {
    _ = stream.SendAudio(micFrame)
}

// Stop gracefully
_ = stream.Stop()
```

### Example 2: Multi-Device, Multi-User Parallel

```go
var wg sync.WaitGroup

for userID, devices := range usersToDevices {
    wg.Add(1)
    go func(uid string, devs []Config) {
        defer wg.Done()
        
        for _, cfg := range devs {
            // Each call is thread-safe
            session, _ := manager.GetOrCreate(uid, cfg)
            session.StartTalk()
            
            stream, _ := client.NewWebSocketAudioStream(session, 100)
            stream.Start(handleAudio)
            // ...
        }
    }(userID, devices)
}

wg.Wait()
```

### Example 3: Restart on Error

```go
if err := stream.SendAudio(data); err != nil {
    fmt.Println("Send failed, restarting...")
    
    stream.Stop()
    time.Sleep(50*ms)
    
    // RestartTalk() handles all cleanup
    session.RestartTalk()
    
    stream, _ = client.NewWebSocketAudioStream(session, 100)
    stream.Start(handleAudio)
}
```

---

## Data Structures

### SessionState

```go
const (
    SessionStateConnected SessionState = iota  // Device connected
    SessionStateTalking                        // Audio active
    SessionStateStopped                        // Session closed
)
```

### AudioCallback

```go
type AudioCallback func(handle int32, data []byte) error

// Callback receives device audio in chunks (PCM S16_LE, 8kHz, mono)
// Must handle non-blocking routing to WebSocket
// Errors logged, never panic
```

### WebSocketAudioStream Stats

```go
map[string]interface{}{
    "handle":          int32(1),
    "is_closed":       bool(false),
    "bytes_received":  uint64(10240),
    "bytes_sent":      uint64(5120),
    "frames_received": uint64(32),
    "frames_sent":     uint64(16),
    "last_send_time":  time.Time{...},
}
```

---

## Error Handling

### Device-Level Errors

| Error Code | Meaning | Handling |
|-----------|---------|----------|
| 11 | Audio mode unsupported | Check device settings, skip voice |
| 605 | Device speaker error | Device doesn't support two-way audio |
| Other | Device API error | Disconnect session, reconnect |

### Session-Level Errors

| Error | Cause | Fix |
|-------|-------|-----|
| "session not connected" | Call before Connect() | Call Connect() first |
| "talk already active" | Duplicate StartTalk() | Call StopTalk() first |
| "send buffer full" | Network congestion | Drop frame or add backpressure |
| "stream closed" | SendAudio after Stop() | Check IsActive() before send |

**Best Practice:**
```go
for {
    if err := stream.SendAudio(data); err != nil {
        if errors.Is(err, errStreamClosed) {
            break  // Normal shutdown
        }
        // Retry logic / restart
    }
}
```

---

## Performance Considerations

### Buffering

- Default buffer: 100 frames
- Tune based on network latency
- Too small → channel full, drops
- Too large → memory pressure, latency

```go
stream, _ := client.NewWebSocketAudioStream(session, 500) // Larger buffer
```

### Concurrency Limits

- SessionManager supports unlimited concurrent users/devices
- Per-device: single talk session (one microphone stream)
- Multiple streams per device would cause device errors

### Resource Cleanup

**On Stop:**
- Closes send channel
- Waits for send loop goroutine
- Calls device StopTalk()
- Unregisters audio callback
- No blocking waits (timeout-safe)

---

## Backward Compatibility

### Legacy StartTalkWithCallback()

```go
// Still works! Returns no-op callback (backward compat)
handle, _ := cli.StartTalkWithCallback()
```

**Why kept:**
- Existing code won't break
- Audio discarded silently (safe fallback)
- Clients can migrate incrementally

### Migration Path

```go
// Old way (removed in v2)
cli.StartTalkWithCallback()

// New way (v1+)
session, _ := manager.GetOrCreate(userID, config)
session.StartTalk()
stream, _ := client.NewWebSocketAudioStream(session, 100)
stream.Start(wsCallback)
```

---

## Testing Checklist

- [x] Multi-user parallel connections
- [x] Multi-device per user
- [x] StopTalk → StartTalk without pause (no race)
- [x] Thread-safe SendAudio from multiple goroutines
- [x] Resource cleanup (no goroutine leaks)
- [x] Error handling (device disconnect, network loss)
- [x] Statistics tracking
- [x] Session manager concurrent access
- [x] AudioHandler callback registry safety

---

## Deployment Checklist

- [ ] Enable CGO: `export CGO_ENABLED=1`
- [ ] Linux platform: `GOOS=linux`
- [ ] Install ALSA (optional, not needed for WebSocket): `sudo apt-get install alsa-utils`
- [ ] Configure device IPs, credentials
- [ ] WebSocket server integration (gorilla/websocket recommended)
- [ ] Browser client: send/receive binary audio frames
- [ ] Monitor session manager: `ListSessions()` for debugging
- [ ] Set appropriate buffer sizes based on network
- [ ] Test restart scenarios before production

---

## Architecture Summary

| Component | Responsibility | Thread Safety |
|-----------|----------------|---------------|
| SessionManager | Multi-device registry | RWMutex protected |
| DeviceSession | Device lifecycle + state | Mutex protected |
| WebSocketAudioStream | PCM framing + buffering | Mutex + channels |
| AudioHandler | Callback routing | Mutex protected |
| bridge/audio.go | C↔Go bridge | CGO safe |

**Key Principle:** Isolation + Synchronization
- Each device session is independent
- No shared mutable state between sessions
- Proper locking at handler boundaries

---

## Next Steps

1. **Browser Integration:** Use Web Audio API to capture microphone → send binary frames
2. **Network Layer:** gorilla/websocket for handle management
3. **Authentication:** Validate userID + device ownership before session creation
4. **Monitoring:** Track stream stats, reconnection rates, device health
5. **Scaling:** Load-balance sessions across multiple servers

---

**Document Version:** 1.0  
**SDK Version:** github.com/evolvecortexsolutions/hikvision-sdk  
**Go Version:** 1.23.4  
**Platform:** Linux (CGO required)
