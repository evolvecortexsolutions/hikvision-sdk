package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/evolvecortexsolutions/hikvision-sdk/client"
)

func main() {
	// NOTE: Replace with your real device credentials and IPs.
	devices := []client.Config{
		{IP: "192.168.1.100", Port: 8000, Username: "admin", Password: "12345"},
		{IP: "192.168.1.101", Port: 8000, Username: "admin", Password: "12345"},
	}

	var wg sync.WaitGroup
	for i, cfg := range devices {
		wg.Add(1)
		go func(index int, cfg client.Config) {
			defer wg.Done()

			cli := client.New(cfg)
			if err := cli.Login(); err != nil {
				fmt.Printf("device %d login failed: %v\n", index, err)
				return
			}

			fmt.Printf("device %d logged in, session=%d\n", index, cli.SessionID())

			// Start talk channel.
			voiceHandle, err := cli.StartTalk()
			if err != nil {
				fmt.Printf("device %d start talk failed: %v\n", index, err)
			} else {
				fmt.Printf("device %d started talk, handle=%d\n", index, voiceHandle)
				if stopErr := cli.StopTalk(voiceHandle); stopErr != nil {
					fmt.Printf("device %d stop talk failed: %v\n", index, stopErr)
				} else {
					fmt.Printf("device %d stopped talk\n", index)
				}
			}

			// Start playback for a 10-second interval (example values).
			start := time.Now().Add(-10 * time.Minute)
			end := time.Now().Add(-9 * time.Minute)
			playHandle, err := cli.StartPlayback(start, end, 0, 0)
			if err != nil {
				fmt.Printf("device %d playback start failed: %v\n", index, err)
			} else {
				fmt.Printf("device %d playback started, handle=%d\n", index, playHandle)
				if stopErr := cli.StopPlayback(playHandle); stopErr != nil {
					fmt.Printf("device %d playback stop failed: %v\n", index, stopErr)
				} else {
					fmt.Printf("device %d playback stopped\n", index)
				}
			}

			// Do something with the session. Minimal example sleep.
			time.Sleep(500 * time.Millisecond)

			if err := cli.Logout(); err != nil {
				fmt.Printf("device %d logout failed: %v\n", index, err)
				return
			}

			fmt.Printf("device %d logged out\n", index)
		}(i, cfg)
	}

	wg.Wait()

	// Global cleanup once all clients are done.
	client.Cleanup()
	fmt.Println("SDK cleanup complete")
}
