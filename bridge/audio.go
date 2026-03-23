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
	"unsafe"
)

//export goAudioCallback
func goAudioCallback(handle C.int, data *C.char, size C.uint, flag C.uchar, user C.uint) {
	goBytes := C.GoBytes(unsafe.Pointer(data), C.int(size))
	handleAudioData(int32(handle), goBytes)
}

// handleAudioData routes device audio to registered handler callbacks.
func handleAudioData(handle int32, data []byte) {
	if err := GetAudioHandler().CallAudio(handle, data); err != nil {
		// Silently ignore if no callback registered (e.g., after stop)
		// This prevents spam when stopping/restarting quickly
	}
}

// StartTalkWithCallback is deprecated: use StartTalkWithWebSocket instead.
// Kept for backward compatibility—callbacks no longer route to aplay.
func StartTalkWithCallback(userID int32) (int32, error) {
	// Default no-op callback (backward compat mode)
	return StartTalkWithWebSocket(userID, func(data []byte) error {
		return nil // silently discard audio
	})
}

// StartTalkWithWebSocket starts voice communication and registers a WebSocket audio callback.
// The callback receives device audio frames and must handle WebSocket transmission.
func StartTalkWithWebSocket(userID int32, audioCallback func([]byte) error) (int32, error) {
	if audioCallback == nil {
		return -1, fmt.Errorf("audio callback cannot be nil")
	}

	handle := C.NET_DVR_StartVoiceCom(
		C.int(userID),
		(C.fVoiceDataCallBack)(unsafe.Pointer(C.goAudioCallback)),
		0,
	)

	if handle == -1 {
		errCode := C.NET_DVR_GetLastError()
		switch errCode {
		case 11:
			return -1, fmt.Errorf("voice intercom not supported by device (audio card mode error)")
		case 605:
			return -1, fmt.Errorf("voice intercom failed with code 605: device does not support speaker audio")
		default:
			return -1, fmt.Errorf("StartVoice failed: %d", errCode)
		}
	}

	// Register the WebSocket callback
	if err := RegisterAudioCallback(int32(handle), audioCallback); err != nil {
		C.NET_DVR_StopVoiceCom(handle)
		return -1, err
	}

	return int32(handle), nil
}

// StopTalkWithWebSocket stops voice communication and unregisters callback.
func StopTalkWithWebSocket(handle int32) error {
	UnregisterAudioCallback(handle)

	ret := C.NET_DVR_StopVoiceCom(C.int(handle))
	if ret == 0 {
		errCode := C.NET_DVR_GetLastError()
		return fmt.Errorf("NET_DVR_StopVoiceCom failed with code %d", int(errCode))
	}

	return nil
}

// SendAudio sends raw PCM data into the active voice talk channel.
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
