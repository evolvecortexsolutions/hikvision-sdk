package bridge

/*
#include <stdlib.h>

typedef void (*fVoiceDataCallBack)(int, char*, unsigned int, unsigned char, unsigned int);

// forward declaration
extern void goAudioCallback(int, char*, unsigned int, unsigned char, unsigned int);
*/
import "C"

import (
	"fmt"
	"os/exec"
	"sync"
	"unsafe"
)

var speakerMu sync.Mutex

//export goAudioCallback
func goAudioCallback(handle C.int, data *C.char, size C.uint, flag C.uchar, user C.uint) {
	goBytes := C.GoBytes(unsafe.Pointer(data), C.int(size))

	// Play received audio
	go playAudio(goBytes)
}

func playAudio(data []byte) {
	speakerMu.Lock()
	defer speakerMu.Unlock()

	cmd := exec.Command("aplay", "-f", "S16_LE", "-r", "8000", "-c", "1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("speaker pipe error:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("speaker start error:", err)
		return
	}

	stdin.Write(data)
	stdin.Close()
	cmd.Wait()
}

// StartTalk with callback
func StartTalkWithCallback(userID int32) (int32, error) {
	handle := C.NET_DVR_StartVoiceCom(
		C.int(userID),
		(C.fVoiceDataCallBack)(unsafe.Pointer(C.goAudioCallback)),
		0,
	)

	if handle == -1 {
		return -1, fmt.Errorf("StartVoice failed: %d", C.NET_DVR_GetLastError())
	}

	return int32(handle), nil
}

// Send audio to device
func SendAudio(handle int32, data []byte) {
	if len(data) == 0 {
		return
	}

	C.NET_DVR_VoiceComSendData(
		C.int(handle),
		(*C.char)(unsafe.Pointer(&data[0])),
		C.uint(len(data)),
	)
}
