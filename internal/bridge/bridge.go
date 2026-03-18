package bridge

// #cgo CFLAGS: -I${SRCDIR}/../../sdk/incEn
// #cgo LDFLAGS: -L${SRCDIR}/../../sdk/lib -Wl,-rpath,${SRCDIR}/../../sdk/lib -lhcnetsdk
/*
#include <stdlib.h>

typedef int BOOL;
typedef unsigned short WORD;
typedef int LONG;
typedef unsigned int DWORD;
typedef unsigned int UINT;
typedef unsigned char BYTE;
typedef void* LPVOID;
typedef void* HWND;

typedef struct {
    char sDVRIP[128];
    WORD wDVRPort;
    char sUserName[32];
    char sPassword[16];
    DWORD dwBaudRate;
    BYTE byDVRType;
    BYTE byAlarmEnable;
    WORD wDecoderDevicePort;
    WORD byChannel;
    BYTE byStartingChannel;
    BYTE byAudioChannel;
    BYTE byIPChannel;
    BYTE byZeroChannel;
    BYTE byMainProto;
    BYTE bySubProto;
    BYTE bySupport;
    DWORD dwSupport;
    char sSupportRet[16];
    DWORD dwSerialNumber;
    DWORD dwAlarmInPortNum;
    DWORD dwAlarmOutPortNum;
    DWORD dwDiskNum;
    BYTE byDVRTypeEx;
    BYTE bySupport2;
    WORD byRes2;
    DWORD dwStep;
    DWORD dwChanNum;
    BYTE byVoiceInChanNum;
    BYTE byStartChan;
    BYTE byRes[2];
} NET_DVR_DEVICEINFO_V30;

typedef struct {
    DWORD dwYear;
    DWORD dwMonth;
    DWORD dwDay;
    DWORD dwHour;
    DWORD dwMinute;
    DWORD dwSecond;
} NET_DVR_TIME;

typedef struct {
    DWORD dwSize;
    void* struIDInfo;
    NET_DVR_TIME struBeginTime;
    NET_DVR_TIME struEndTime;
    HWND hWnd;
    BYTE byDrawFrame;
    BYTE byVolumeType;
    BYTE byVolumeNum;
    BYTE byStreamType;
    DWORD dwFileIndex;
    BYTE byAudioFile;
    BYTE byCourseFile;
    BYTE byDownload;
    BYTE byOptimalStreamType;
    BYTE byUseAsyn;
    BYTE byRes2[19];
} NET_DVR_VOD_PARA;

extern BOOL NET_DVR_Init(void);
extern BOOL NET_DVR_Cleanup(void);
extern LONG NET_DVR_Login_V30(char *sDVRIP, WORD wDVRPort, char *sUserName, char *sPassword, NET_DVR_DEVICEINFO_V30 *lpDeviceInfo);
extern BOOL NET_DVR_Logout(LONG lUserID);
extern DWORD NET_DVR_GetLastError(void);
extern LONG NET_DVR_StartVoiceCom(LONG lUserID, void(*)(LONG, char*, DWORD, BYTE, DWORD), DWORD dwUser);
extern BOOL NET_DVR_StopVoiceCom(LONG lVoiceComHandle);
extern LONG NET_DVR_PlayBackByTime_V40(LONG lUserID, const NET_DVR_VOD_PARA* pVodPara);
extern BOOL NET_DVR_StopPlayBack(LONG lPlayHandle);
*/
import "C"

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
	userID := C.NET_DVR_Login_V30(cIP, C.WORD(port), cUser, cPass, &info)
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
	if C.NET_DVR_Logout(C.LONG(userID)) == 0 {
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
	voiceHandle := C.NET_DVR_StartVoiceCom(C.LONG(userID), nil, 0)
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
	if C.NET_DVR_StopVoiceCom(C.LONG(voiceHandle)) == 0 {
		errCode := C.NET_DVR_GetLastError()
		return fmt.Errorf("NET_DVR_StopVoiceCom failed with code %d", int(errCode))
	}
	return nil
}

// PlayBackByTime starts a VOD session and returns playback handle.
func PlayBackByTime(userID int32, start time.Time, end time.Time, streamType uint8, fileIndex uint32) (int32, error) {
	if !inited {
		return -1, errors.New("SDK not initialized: call InitSDK first")
	}

	var vod C.NET_DVR_VOD_PARA
	vod.dwSize = C.DWORD(unsafe.Sizeof(vod))
	vod.struBeginTime.dwYear = C.DWORD(start.Year())
	vod.struBeginTime.dwMonth = C.DWORD(start.Month())
	vod.struBeginTime.dwDay = C.DWORD(start.Day())
	vod.struBeginTime.dwHour = C.DWORD(start.Hour())
	vod.struBeginTime.dwMinute = C.DWORD(start.Minute())
	vod.struBeginTime.dwSecond = C.DWORD(start.Second())
	vod.struEndTime.dwYear = C.DWORD(end.Year())
	vod.struEndTime.dwMonth = C.DWORD(end.Month())
	vod.struEndTime.dwDay = C.DWORD(end.Day())
	vod.struEndTime.dwHour = C.DWORD(end.Hour())
	vod.struEndTime.dwMinute = C.DWORD(end.Minute())
	vod.struEndTime.dwSecond = C.DWORD(end.Second())
	vod.hWnd = C.HWND(nil)
	vod.byDrawFrame = 0
	vod.byVolumeType = 0
	vod.byVolumeNum = 0
	vod.byStreamType = C.BYTE(streamType)
	vod.dwFileIndex = C.DWORD(fileIndex)
	vod.byAudioFile = 0
	vod.byCourseFile = 0
	vod.byDownload = 0
	vod.byOptimalStreamType = 0
	vod.byUseAsyn = 0

	playHandle := C.NET_DVR_PlayBackByTime_V40(C.LONG(userID), &vod)
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
	if C.NET_DVR_StopPlayBack(C.LONG(playHandle)) == 0 {
		errCode := C.NET_DVR_GetLastError()
		return fmt.Errorf("NET_DVR_StopPlayBack failed with code %d", int(errCode))
	}
	return nil
}
