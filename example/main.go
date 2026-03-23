//go:build linux

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

			fmt.Printf("device %d logged in, session=%d, type=%v\n", index, cli.SessionID(), cfg.DeviceType)

			if err := cli.StartKeepAlive(20 * time.Second); err != nil {
				fmt.Printf("device %d keepalive failed: %v\n", index, err)
			}

			voiceHandle, err := cli.StartTalkWithCallback()
			if err != nil {
				fmt.Printf("device %d start talk failed: %v\n", index, err)
			} else {
				fmt.Printf("device %d started talk, handle=%d\n", index, voiceHandle)
				if err := cli.SendAudio(voiceHandle, make([]byte, 320)); err != nil {
					fmt.Printf("device %d send audio failed: %v\n", index, err)
				}
				time.Sleep(2 * time.Second)
				if err := cli.StopTalk(voiceHandle); err != nil {
					fmt.Printf("device %d stop talk failed: %v\n", index, err)
				}
			}

			if state, err := cli.GetDVRWorkState(); err != nil {
				fmt.Printf("device %d work state read failed: %v\n", index, err)
			} else {
				fmt.Printf("device %d work state: %+v\n", index, state)
			}

			if networks, err := cli.GetNetworkConfig(); err != nil {
				fmt.Printf("device %d network config read failed: %v\n", index, err)
			} else {
				for i, net := range networks {
					fmt.Printf("device %d network %d: IP=%s, MAC=%s, Gateway=%s, DHCP=%v\n",
						index, i, net.IP, net.MAC, net.Gateway, net.DHCP)
				}
			}

			start := time.Now().Add(-10 * time.Minute)
			end := time.Now().Add(-9 * time.Minute)
			if playHandle, err := cli.StartPlayback(start, end, 0, 0); err != nil {
				fmt.Printf("device %d playback start failed: %v\n", index, err)
			} else {
				fmt.Printf("device %d playback started, handle=%d\n", index, playHandle)
				if err := cli.StopPlayback(playHandle); err != nil {
					fmt.Printf("device %d playback stop failed: %v\n", index, err)
				}
			}

			cli.StopKeepAlive()
			if err := cli.Logout(); err != nil {
				fmt.Printf("device %d logout failed: %v\n", index, err)
			}
			fmt.Printf("device %d logged out\n", index)
		}(i, cfg)
	}

	wg.Wait()

	// Global cleanup once all clients are done.
	client.Cleanup()
	fmt.Println("SDK cleanup complete")
}
