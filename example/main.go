package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/evolvecortexsolutions/hikvision-go-wrapper/client"
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
