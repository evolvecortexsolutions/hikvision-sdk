package bridge

// #cgo CFLAGS: -I${SRCDIR}/../../sdk/incEn
// #cgo LDFLAGS: -L${SRCDIR}/../../sdk/lib -Wl,-rpath,${SRCDIR}/../../sdk/lib -lhcnetsdk
// #include <stdlib.h>
//
// typedef int BOOL;
// typedef unsigned short WORD;
// typedef int LONG;
// typedef unsigned int DWORD;
// typedef void* LPVOID;
// typedef void* LPNET_DVR_DEVICEINFO_V30;
//
// typedef struct {
//     unsigned char _reserved[1024];
// } NET_DVR_DEVICEINFO_V30;
//
// extern BOOL NET_DVR_Init(void);
// extern BOOL NET_DVR_Cleanup(void);
// extern LONG NET_DVR_Login_V30(char *sDVRIP, WORD wDVRPort, char *sUserName, char *sPassword, NET_DVR_DEVICEINFO_V30 *lpDeviceInfo);
// extern BOOL NET_DVR_Logout(LONG lUserID);
// extern DWORD NET_DVR_GetLastError(void);
import "C"

import (
	"errors"
	"fmt"
	"sync"
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
