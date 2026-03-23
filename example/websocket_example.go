//go:build linux

package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/evolvecortexsolutions/hikvision-sdk/client"
)

// Example: Multi-device, multi-user with WebSocket audio
func main() {
	// Initialize session manager (global for your server)
	manager := client.NewSessionManager()
	defer manager.CloseAllSessions()

	// Scenario: Two users, each connecting to different devices
	users := []struct {
		userID  string
		devices []client.Config
	}{
		{
			userID: "user_alice",
			devices: []client.Config{
				{IP: "192.168.1.100", Port: 8000, Username: "admin", Password: "12345"},
				{IP: "192.168.1.101", Port: 8000, Username: "admin", Password: "12345"},
			},
		},
		{
			userID: "user_bob",
			devices: []client.Config{
				{IP: "192.168.1.102", Port: 8000, Username: "admin", Password: "12345"},
			},
		},
	}

	var wg sync.WaitGroup

	// Process each user independently
	for _, user := range users {
		wg.Add(1)
		go func(userID string, devices []client.Config) {
			defer wg.Done()

			fmt.Printf("\n=== User %s ===\n", userID)

			var deviceWg sync.WaitGroup

			// Each user manages multiple devices in parallel
			for i, cfg := range devices {
				deviceWg.Add(1)
				go func(deviceIdx int, deviceCfg client.Config) {
					defer deviceWg.Done()

					fmt.Printf("[%s Device %d] Connecting to %s:%d\n", userID, deviceIdx, deviceCfg.IP, deviceCfg.Port)

					// Get or create session (thread-safe)
					session, err := manager.GetOrCreate(userID, deviceCfg)
					if err != nil {
						fmt.Printf("[%s Device %d] Connection failed: %v\n", userID, deviceIdx, err)
						return
					}

					fmt.Printf("[%s Device %d] Connected, state=%s\n", userID, deviceIdx, session.State())

					// Simulate audio session with WebSocket
					if err := demonstrateAudioSession(userID, deviceIdx, session); err != nil {
						fmt.Printf("[%s Device %d] Audio session error: %v\n", userID, deviceIdx, err)
					}

					// Print session stats
					stats := session.GetStats()
					fmt.Printf("[%s Device %d] Session stats: %+v\n", userID, deviceIdx, stats)

				}(i, cfg)
			}

			deviceWg.Wait()
			fmt.Printf("[%s] All devices completed\n", userID)

		}(user.userID, user.devices)
	}

	wg.Wait()

	// Print final manager state
	fmt.Printf("\n=== Final Session Manager State ===\n")
	allSessions := manager.ListSessions()
	for key, stats := range allSessions {
		fmt.Printf("Session %s: %+v\n", key, stats)
	}

	client.Cleanup()
	fmt.Println("\nSDK cleanup complete")
}

// demonstrateAudioSession shows a complete audio session lifecycle
func demonstrateAudioSession(userID string, deviceIdx int, session *client.DeviceSession) error {
	prefix := fmt.Sprintf("[%s Device %d]", userID, deviceIdx)

	// ===== Test 1: Start Talk =====
	fmt.Printf("%s Starting audio session...\n", prefix)
	if err := session.StartTalk(); err != nil {
		return fmt.Errorf("start talk failed: %w", err)
	}
	fmt.Printf("%s Talk started, state=%s\n", prefix, session.State())

	// Create WebSocket audio stream
	stream, err := client.NewWebSocketAudioStream(session, 100)
	if err != nil {
		return fmt.Errorf("create stream failed: %w", err)
	}

	// Start stream with mock WebSocket callback
	mockWebSocketReceiver := func(data []byte) error {
		fmt.Printf("%s [WebSocket RX] Received %d bytes\n", prefix, len(data))
		// In production: send via ws.WriteMessage() to browser
		return nil
	}

	if err := stream.Start(mockWebSocketReceiver); err != nil {
		return fmt.Errorf("stream start failed: %w", err)
	}
	fmt.Printf("%s Audio stream active\n", prefix)

	// Simulate browser sending audio (microphone data)
	mockMicData := make([]byte, 320)
	for i := 0; i < 3; i++ {
		if err := stream.SendAudio(mockMicData); err != nil {
			fmt.Printf("%s Send audio error: %v\n", prefix, err)
		} else {
			fmt.Printf("%s [WebSocket TX] Sent 320 bytes\n", prefix)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// ===== Test 2: Quick Restart (Stop → Start) =====
	fmt.Printf("%s Testing restart...\n", prefix)
	if err := stream.Stop(); err != nil {
		return fmt.Errorf("stream stop failed: %w", err)
	}
	fmt.Printf("%s Stream stopped\n", prefix)

	time.Sleep(50 * time.Millisecond)

	// Immediately restart
	stream, err = client.NewWebSocketAudioStream(session, 100)
	if err != nil {
		return fmt.Errorf("recreate stream failed: %w", err)
	}

	if err := stream.Start(mockWebSocketReceiver); err != nil {
		return fmt.Errorf("restart stream failed: %w", err)
	}
	fmt.Printf("%s Stream restarted successfully\n", prefix)

	// Send more data on restarted stream
	if err := stream.SendAudio(mockMicData); err != nil {
		fmt.Printf("%s Send audio on restarted stream error: %v\n", prefix, err)
	} else {
		fmt.Printf("%s [WebSocket TX] Sent 320 bytes on restarted stream\n", prefix)
	}

	// Print stream stats
	streamStats := stream.Stats()
	fmt.Printf("%s Stream stats: %+v\n", prefix, streamStats)

	// Clean stop
	if err := stream.Stop(); err != nil {
		return fmt.Errorf("stream stop failed: %w", err)
	}

	fmt.Printf("%s Audio session completed successfully\n", prefix)
	return nil
}

// Example: Parallel device operations with restart scenarios
func exampleParallelRestart() {
	manager := client.NewSessionManager()
	defer manager.CloseAllSessions()

	cfg := client.Config{
		IP:       "192.168.1.100",
		Port:     8000,
		Username: "admin",
		Password: "12345",
	}

	session, _ := manager.GetOrCreate("user1", cfg)

	session.StartTalk()
	fmt.Println("Talk started")

	time.Sleep(100 * time.Millisecond)

	// Immediate restart (critical race condition test)
	session.RestartTalk()
	fmt.Println("Talk restarted")

	time.Sleep(100 * time.Millisecond)

	// Another restart
	session.RestartTalk()
	fmt.Println("Talk restarted again")

	session.Disconnect()
}

// Example: Streaming with timeout and recovery
func exampleStreamingWithRecovery() {
	manager := client.NewSessionManager()
	defer manager.CloseAllSessions()

	cfg := client.Config{
		IP:       "192.168.1.100",
		Port:     8000,
		Username: "admin",
		Password: "12345",
	}

	session, _ := manager.GetOrCreate("user1", cfg)
	session.StartTalk()

	stream, _ := client.NewWebSocketAudioStream(session, 100)
	stream.Start(func(data []byte) error { return nil })

	// Simulate streaming
	for i := 0; i < 5; i++ {
		data := make([]byte, 320)
		if err := stream.SendAudio(data); err != nil {
			fmt.Printf("Send error: %v, restarting...\n", err)

			stream.Stop()
			time.Sleep(50 * time.Millisecond)

			stream, _ = client.NewWebSocketAudioStream(session, 100)
			stream.Start(func(data []byte) error { return nil })
		}
		time.Sleep(100 * time.Millisecond)
	}

	stream.Stop()
	session.Disconnect()
}
