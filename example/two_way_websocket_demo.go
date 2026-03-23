//go:build linux

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/evolvecortexsolutions/hikvision-sdk/client"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

// TwoWayWebSocketDemo demonstrates complete 2-way audio communication
// Server side: Go with Hikvision SDK + WebSocket server
// Client side: Browser with WebRTC microphone + WebSocket
func main() {
	// Initialize session manager
	manager := client.NewSessionManager()
	defer manager.CloseAllSessions()

	// Device configuration
	cfg := client.Config{
		IP:       "192.168.1.100", // Change to your device IP
		Port:     8000,
		Username: "admin",
		Password: "12345",
	}

	// Get device session
	session, err := manager.GetOrCreate("demo_user", cfg)
	if err != nil {
		log.Fatal("Failed to connect to device:", err)
	}

	// Create WebSocket audio stream
	stream, err := client.NewWebSocketAudioStream(session, 100)
	if err != nil {
		log.Fatal("Failed to create audio stream:", err)
	}

	// WebSocket handler for browser connections
	http.HandleFunc("/audio", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade failed:", err)
			return
		}
		defer conn.Close()

		fmt.Println("Browser connected via WebSocket")

		// Start audio stream with WebSocket sender
		webSocketSender := func(data []byte) error {
			// Send audio data from device to browser
			return conn.WriteMessage(websocket.BinaryMessage, data)
		}

		if err := stream.Start(webSocketSender); err != nil {
			log.Println("Failed to start stream:", err)
			return
		}

		// Handle incoming messages from browser (microphone audio)
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket read error:", err)
				break
			}

			if messageType == websocket.BinaryMessage {
				// Send microphone audio to device
				if err := stream.SendAudio(data); err != nil {
					log.Println("Failed to send audio to device:", err)
				}
			}
		}

		// Stop stream when connection closes
		stream.Stop()
	})

	// Serve HTML page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Hikvision 2-Way Audio Demo</title>
</head>
<body>
    <h1>Hikvision 2-Way Audio Demo</h1>
    <button id="startBtn">Start Audio</button>
    <button id="stopBtn">Stop Audio</button>
    <div id="status">Status: Disconnected</div>

    <script>
        let ws;
        let mediaRecorder;
        let audioContext;
        let analyser;
        let microphone;

        const startBtn = document.getElementById('startBtn');
        const stopBtn = document.getElementById('stopBtn');
        const status = document.getElementById('status');

        startBtn.onclick = async () => {
            try {
                // Connect WebSocket
                ws = new WebSocket('ws://localhost:8080/audio');
                
                ws.onopen = () => {
                    status.textContent = 'Status: Connected';
                    startMicrophone();
                };

                ws.onmessage = (event) => {
                    // Play received audio from device
                    playAudio(event.data);
                };

                ws.onclose = () => {
                    status.textContent = 'Status: Disconnected';
                    stopMicrophone();
                };

                ws.onerror = (error) => {
                    status.textContent = 'Status: Error - ' + error;
                };

            } catch (error) {
                status.textContent = 'Status: Error - ' + error.message;
            }
        };

        stopBtn.onclick = () => {
            if (ws) {
                ws.close();
            }
            stopMicrophone();
        };

        async function startMicrophone() {
            try {
                // Get microphone access
                const stream = await navigator.mediaDevices.getUserMedia({ 
                    audio: {
                        sampleRate: 8000,
                        channelCount: 1,
                        echoCancellation: true,
                        noiseSuppression: true
                    }
                });

                audioContext = new AudioContext({ sampleRate: 8000 });
                analyser = audioContext.createAnalyser();
                microphone = audioContext.createMediaStreamSource(stream);
                microphone.connect(analyser);

                // Record and send audio chunks
                mediaRecorder = new MediaRecorder(stream, {
                    mimeType: 'audio/webm;codecs=pcm'
                });

                mediaRecorder.ondataavailable = (event) => {
                    if (event.data.size > 0 && ws.readyState === WebSocket.OPEN) {
                        // Convert to PCM and send
                        event.data.arrayBuffer().then(buffer => {
                            const pcmData = convertToPCM(new Uint8Array(buffer));
                            ws.send(pcmData);
                        });
                    }
                };

                mediaRecorder.start(100); // Send every 100ms
                status.textContent = 'Status: Connected + Microphone Active';

            } catch (error) {
                status.textContent = 'Status: Microphone Error - ' + error.message;
            }
        }

        function stopMicrophone() {
            if (mediaRecorder && mediaRecorder.state !== 'inactive') {
                mediaRecorder.stop();
            }
            if (microphone) {
                microphone.disconnect();
            }
            if (audioContext) {
                audioContext.close();
            }
        }

        function playAudio(audioData) {
            // Create audio buffer and play
            const audioBuffer = audioContext.createBuffer(1, audioData.length / 2, 8000);
            const channelData = audioBuffer.getChannelData(0);
            
            // Convert PCM S16_LE to float32
            for (let i = 0; i < audioData.length; i += 2) {
                const sample = (audioData[i+1] << 8) | audioData[i];
                channelData[i/2] = sample / 32768.0;
            }

            const source = audioContext.createBufferSource();
            source.buffer = audioBuffer;
            source.connect(audioContext.destination);
            source.start();
        }

        function convertToPCM(webmData) {
            // Simple conversion - in production, use proper WebM to PCM conversion
            // For demo, assume WebM contains PCM data
            return webmData;
        }
    </script>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	fmt.Println("Starting 2-way audio demo server on :8080")
	fmt.Println("Open http://localhost:8080 in your browser")
	fmt.Println("Make sure your Hikvision device is accessible at", cfg.IP)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
