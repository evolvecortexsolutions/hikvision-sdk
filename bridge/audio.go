package bridge

/*
#include <stdlib.h>

typedef void (*fVoiceDataCallBack)(int, char*, unsigned int, unsigned char, unsigned int);

extern int NET_DVR_StartVoiceCom(int lUserID, fVoiceDataCallBack callback, unsigned int dwUser);
extern int NET_DVR_StopVoiceCom(int lVoiceComHandle);
extern int NET_DVR_VoiceComSendData(int lVoiceComHandle, char *pSendBuf, unsigned int dwBufSize);
extern unsigned int NET_DVR_GetLastError(void);

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
	playAudio(goBytes)
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

func StartTalkWithCallback(userID int32) (int32, error) {
	handle := C.NET_DVR_StartVoiceCom(
		C.int(userID),
		(C.fVoiceDataCallBack)(unsafe.Pointer(C.goAudioCallback)),
		0,
	)

	if handle == -1 {
		errCode := C.NET_DVR_GetLastError()
		if errCode == 11 { // NET_DVR_AUDIO_MODE_ERROR
			return -1, fmt.Errorf("voice intercom not supported by device (audio card mode error)")
		}
		return -1, fmt.Errorf("StartVoice failed: %d", errCode)
	}

	return int32(handle), nil
}

func SendAudio(handle int32, data []byte) error {
	if handle < 0 {
		return fmt.Errorf("invalid voice handle")
	}
	if len(data) == 0 {
		return fmt.Errorf("payload cannot be empty")
	}

	ret := C.NET_DVR_VoiceComSendData(
		C.int(handle),
		(*C.char)(unsafe.Pointer(&data[0])),
		C.uint(len(data)),
	)
	if ret == 0 {
		return fmt.Errorf("NET_DVR_VoiceComSendData failed with code %d", int(C.NET_DVR_GetLastError()))
	}
	return nil
}
