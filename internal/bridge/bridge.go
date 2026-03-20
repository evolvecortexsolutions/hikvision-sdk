package bridge

// #cgo CFLAGS: -I${SRCDIR}/../../sdk/incEn
// #cgo LDFLAGS: -L${SRCDIR}/../../sdk/lib -Wl,-rpath,${SRCDIR}/../../sdk/lib -lhcnetsdk -lopenal -lAudioRender
/*
#include <stdlib.h>

typedef struct {
    char sDVRIP[128];
    unsigned short wDVRPort;
    char sUserName[32];
    char sPassword[16];
    unsigned int dwBaudRate;
    unsigned char byDVRType;
    unsigned char byAlarmEnable;
    unsigned short wDecoderDevicePort;
    unsigned short byChannel;
    unsigned char byStartingChannel;
    unsigned char byAudioChannel;
    unsigned char byIPChannel;
    unsigned char byZeroChannel;
    unsigned char byMainProto;
    unsigned char bySubProto;
    unsigned char bySupport;
    unsigned int dwSupport;
    char sSupportRet[16];
    unsigned int dwSerialNumber;
    unsigned int dwAlarmInPortNum;
    unsigned int dwAlarmOutPortNum;
    unsigned int dwDiskNum;
    unsigned char byDVRTypeEx;
    unsigned char bySupport2;
    unsigned short byRes2;
    unsigned int dwStep;
    unsigned int dwChanNum;
    unsigned char byVoiceInChanNum;
    unsigned char byStartChan;
    unsigned char byRes[2];
} NET_DVR_DEVICEINFO_V30;

typedef struct {
    unsigned int dwYear;
    unsigned int dwMonth;
    unsigned int dwDay;
    unsigned int dwHour;
    unsigned int dwMinute;
    unsigned int dwSecond;
} NET_DVR_TIME;

typedef struct {
    unsigned int dwSize;
    void* struIDInfo;
    NET_DVR_TIME struBeginTime;
    NET_DVR_TIME struEndTime;
    void* hWnd;
    unsigned char byDrawFrame;
    unsigned char byVolumeType;
    unsigned char byVolumeNum;
    unsigned char byStreamType;
    unsigned int dwFileIndex;
    unsigned char byAudioFile;
    unsigned char byCourseFile;
    unsigned char byDownload;
    unsigned char byOptimalStreamType;
    unsigned char byUseAsyn;
    unsigned char byRes2[19];
} NET_DVR_VOD_PARA;

extern int NET_DVR_Init(void);
extern int NET_DVR_Cleanup(void);
extern int NET_DVR_Login_V30(char *sDVRIP, unsigned short wDVRPort, char *sUserName, char *sPassword, NET_DVR_DEVICEINFO_V30 *lpDeviceInfo);
extern int NET_DVR_Logout(int lUserID);
extern unsigned int NET_DVR_GetLastError(void);
extern int NET_DVR_StartVoiceCom(int lUserID, void(*callback)(int, char*, unsigned int, unsigned char, unsigned int), unsigned int dwUser);
extern int NET_DVR_StopVoiceCom(int lVoiceComHandle);
extern int NET_DVR_VoiceComSendData(int lVoiceComHandle, char *pSendBuf, unsigned int dwBufSize);
extern int NET_DVR_PlayBackByTime_V40(int lUserID, const NET_DVR_VOD_PARA* pVodPara);
extern int NET_DVR_StopPlayBack(int lPlayHandle);
*/
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"
)

var (
	initOnce sync.Once
	inited   bool
	initErr  error
	mu       sync.Mutex
)

// InitSDK initializes the underlying HCNetSDK once.
func InitSDK() error {
	initOnce.Do(func() {
		if C.NET_DVR_Init() == 0 {
			initErr = errors.New("NET_DVR_Init failed")
			return
		}
		inited = true
	})
	return initErr
}

// CleanupSDK releases SDK resources.
func CleanupSDK() {
	mu.Lock()
	defer mu.Unlock()
	if !inited {
		return
	}
	C.NET_DVR_Cleanup()
	inited = false
}

// LoginV30 logs into a device and returns a session user ID.
func LoginV30(ip string, port uint16, username string, password string) (int32, error) {
	if !inited {
		return -1, errors.New("SDK not initialized: call InitSDK first")
	}

	cIP := C.CString(ip)
	defer C.free(unsafe.Pointer(cIP))
	cUser := C.CString(username)
	defer C.free(unsafe.Pointer(cUser))
	cPass := C.CString(password)
	defer C.free(unsafe.Pointer(cPass))

	var info C.NET_DVR_DEVICEINFO_V30
	userID := C.NET_DVR_Login_V30(cIP, C.ushort(port), cUser, cPass, &info)
	if userID == -1 {
		errCode := C.NET_DVR_GetLastError()
		return -1, fmt.Errorf("NET_DVR_Login_V30 failed with code %d", int(errCode))
	}
	return int32(userID), nil
}

// Logout logs out a session.
func Logout(userID int32) error {
	if userID < 0 {
		return errors.New("invalid userID")
	}
	if C.NET_DVR_Logout(C.int(userID)) == 0 {
		errCode := C.NET_DVR_GetLastError()
		return fmt.Errorf("NET_DVR_Logout failed with code %d", int(errCode))
	}
	return nil
}

// StartTalk opens a voice talk channel and returns voice handle.
func StartTalk(userID int32) (int32, error) {
	if !inited {
		return -1, errors.New("SDK not initialized: call InitSDK first")
	}
	voiceHandle := C.NET_DVR_StartVoiceCom(C.int(userID), nil, 0)
	if voiceHandle == -1 {
		errCode := C.NET_DVR_GetLastError()
		return -1, fmt.Errorf("NET_DVR_StartVoiceCom failed with code %d", int(errCode))
	}
	return int32(voiceHandle), nil
}

// StopTalk stops voice talk by handle.
func StopTalk(voiceHandle int32) error {
	if voiceHandle < 0 {
		return errors.New("invalid voice handle")
	}
	if C.NET_DVR_StopVoiceCom(C.int(voiceHandle)) == 0 {
		errCode := C.NET_DVR_GetLastError()
		return fmt.Errorf("NET_DVR_StopVoiceCom failed with code %d", int(errCode))
	}
	return nil
}

// SendAudio writes raw PCM bytes to the voice talk channel.
// PlayBackByTime starts a VOD session and returns playback handle.
func PlayBackByTime(userID int32, start time.Time, end time.Time, streamType uint8, fileIndex uint32) (int32, error) {
	if !inited {
		return -1, errors.New("SDK not initialized: call InitSDK first")
	}

	var vod C.NET_DVR_VOD_PARA
	vod.dwSize = C.uint(unsafe.Sizeof(vod))
	vod.struBeginTime.dwYear = C.uint(start.Year())
	vod.struBeginTime.dwMonth = C.uint(start.Month())
	vod.struBeginTime.dwDay = C.uint(start.Day())
	vod.struBeginTime.dwHour = C.uint(start.Hour())
	vod.struBeginTime.dwMinute = C.uint(start.Minute())
	vod.struBeginTime.dwSecond = C.uint(start.Second())
	vod.struEndTime.dwYear = C.uint(end.Year())
	vod.struEndTime.dwMonth = C.uint(end.Month())
	vod.struEndTime.dwDay = C.uint(end.Day())
	vod.struEndTime.dwHour = C.uint(end.Hour())
	vod.struEndTime.dwMinute = C.uint(end.Minute())
	vod.struEndTime.dwSecond = C.uint(end.Second())
	vod.hWnd = nil
	vod.byDrawFrame = C.uchar(0)
	vod.byVolumeType = C.uchar(0)
	vod.byVolumeNum = C.uchar(0)
	vod.byStreamType = C.uchar(streamType)
	vod.dwFileIndex = C.uint(fileIndex)
	vod.byAudioFile = C.uchar(0)
	vod.byCourseFile = C.uchar(0)
	vod.byDownload = C.uchar(0)
	vod.byOptimalStreamType = C.uchar(0)
	vod.byUseAsyn = C.uchar(0)

	playHandle := C.NET_DVR_PlayBackByTime_V40(C.int(userID), &vod)
	if playHandle == -1 {
		errCode := C.NET_DVR_GetLastError()
		return -1, fmt.Errorf("NET_DVR_PlayBackByTime_V40 failed with code %d", int(errCode))
	}
	return int32(playHandle), nil
}

// StopPlayback stops a playback session.
func StopPlayback(playHandle int32) error {
	if playHandle < 0 {
		return errors.New("invalid playback handle")
	}
	if C.NET_DVR_StopPlayBack(C.int(playHandle)) == 0 {
		errCode := C.NET_DVR_GetLastError()
		return fmt.Errorf("NET_DVR_StopPlayBack failed with code %d", int(errCode))
	}
	return nil
}
