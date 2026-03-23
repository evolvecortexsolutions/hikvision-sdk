//go:build linux

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/evolvecortexsolutions/hikvision-sdk/client"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
var sessions = map[string]*client.DeviceSession{}
var sessionsMu sync.Mutex

func newSessionID() string {
	return strconv.FormatInt(int64(rand.Int31()), 10)
}

func main() {
	manager := client.NewSessionManager()
	defer manager.CloseAllSessions()

	cfg := client.Config{IP: "10.0.0.196", Port: 8000, Username: "admin", Password: "12345"}

	http.HandleFunc("/talk/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionID := newSessionID()
		session, err := manager.GetOrCreate(sessionID, cfg)
		if err != nil {
			log.Println("talk/start error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sessionsMu.Lock()
		sessions[sessionID] = session
		sessionsMu.Unlock()

		json.NewEncoder(w).Encode(map[string]string{"session_id": sessionID})
	})

	http.HandleFunc("/talk", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("session_id")
		if id == "" {
			http.Error(w, "missing session_id", http.StatusBadRequest)
			return
		}

		mode := r.URL.Query().Get("mode")
		if mode == "" {
			mode = "both" // both/send-only/receive-only
		}

		sessionsMu.Lock()
		session, ok := sessions[id]
		sessionsMu.Unlock()
		if !ok {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("websocket upgrade failed:", err)
			return
		}
		defer conn.Close()

		var audioStream *client.WebSocketAudioStream

		// Setup receive path (device -> browser)
		if mode != "send-only" {
			audioStream, err = client.NewWebSocketAudioStream(session, 200)
			if err != nil {
				log.Println("NewWebSocketAudioStream error:", err)
				return
			}

			cb := func(devicePCM []byte) error {
				if len(devicePCM)%2 != 0 {
					return nil
				}
				ulaw := make([]byte, len(devicePCM)/2)
				for i := 0; i < len(devicePCM); i += 2 {
					pcm := int16(devicePCM[i]) | int16(devicePCM[i+1])<<8
					ulaw[i/2] = linearToMuLaw(pcm)
				}
				return conn.WriteMessage(websocket.BinaryMessage, ulaw)
			}

			err = audioStream.Start(cb)
			if err != nil {
				log.Println("audioStream.Start error:", err)
				_ = conn.WriteJSON(map[string]string{"type": "ERROR", "msg": err.Error()})
				if !strings.Contains(err.Error(), "605") {
					return
				}
				log.Println("Device does not support speaker audio; continuing in send-only path")
			}
		}

		// Setup send-only path (browser -> device)
		if mode != "receive-only" {
			if err := session.StartTalkSendOnly(); err != nil {
				log.Println("start send-only talk failed:", err)
				_ = conn.WriteJSON(map[string]string{"type": "ERROR", "msg": err.Error()})
				if mode == "send-only" {
					return
				}
			}
		}

		_ = conn.WriteJSON(map[string]string{"type": "TALK_READY", "session_id": id, "mode": mode})

		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("read error", err)
				break
			}

			if mt == websocket.TextMessage {
				var msgObj map[string]interface{}
				if err := json.Unmarshal(msg, &msgObj); err == nil {
					if msgObj["type"] == "SUBSCRIBE_AUDIO" {
						log.Println("SUBSCRIBE_AUDIO received")
					}
				}
				continue
			}

			if mt == websocket.BinaryMessage && mode != "receive-only" {
				pcm := make([]byte, len(msg)*2)
				for i := 0; i < len(msg); i++ {
					val := muLawToLinear(msg[i])
					pcm[2*i] = byte(val)
					pcm[2*i+1] = byte(val >> 8)
				}

				if mode == "send-only" {
					if err := session.SendAudio(pcm); err != nil {
						log.Println("session.SendAudio error", err)
					}
				} else {
					if audioStream != nil {
						if err := audioStream.SendAudio(pcm); err != nil {
							log.Println("audioStream.SendAudio error", err)
						}
					}
				}
			}
		}

		if mode != "send-only" && audioStream != nil {
			audioStream.Stop()
		}

		if mode != "receive-only" {
			_ = session.StopTalk()
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./example/client.html")
	})

	fmt.Println("Starting 2-way audio demo server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func linearToMuLaw(sample int16) byte {
	const clip = 32635
	sign := byte(0)
	if sample < 0 {
		sample = -sample
		sign = 0x80
	}
	if sample > clip {
		sample = clip
	}

	sample = sample + 0x84
	exponent := 7
	mantissa := 0

	for expMask := int16(0x4000); expMask > 0 && (sample&expMask) == 0; expMask >>= 1 {
		exponent--
	}
	mantissa = int((sample >> (exponent + 3)) & 0x0F)

	ulawByte := ^(sign | byte(exponent<<4) | byte(mantissa))
	return ulawByte
}

func muLawToLinear(ulaw byte) int16 {
	ulaw = ^ulaw
	sign := int16(ulaw & 0x80)
	exponent := (ulaw >> 4) & 0x07
	mantissa := ulaw & 0x0F

	sample := int16((uint16(mantissa) << 3) + 0x84)
	sample <<= exponent
	sample -= 0x84

	if sign != 0 {
		sample = -sample
	}
	return sample
}
