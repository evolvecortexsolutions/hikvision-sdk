package bridge

// #cgo CFLAGS: -I${SRCDIR}/../sdk/incEn
// #cgo LDFLAGS: -L${SRCDIR}/../sdk/lib -Wl,-rpath,${SRCDIR}/../sdk/lib -lhcnetsdk -lopenal -lAudioRender
/*
#include <stdlib.h>
#include <string.h>

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

typedef struct tagNET_DVR_DEVICEINFO_V40
{
    NET_DVR_DEVICEINFO_V30 struDeviceV30;
    unsigned char bySupportLock;
    unsigned char byRetryLoginTime;
    unsigned char byPasswordLevel;
    unsigned char byProxyType;
    unsigned int dwSurplusLockTime;
    unsigned char byCharEncodeType;
    unsigned char bySupportDev5;
    unsigned char bySupport;
    unsigned char byLoginMode;
    unsigned int dwOEMCode;
    int iResidualValidity;
    unsigned char byResidualValidity;
    unsigned char bySingleStartDTalkChan;
    unsigned char bySingleDTalkChanNums;
    unsigned char byPassWordResetLevel;
    unsigned char bySupportStreamEncrypt;
    unsigned char byMarketType;
    unsigned char byRes2[238];
} NET_DVR_DEVICEINFO_V40;

typedef struct
{
    char sDeviceAddress[129];
    unsigned char byUseTransport;
    unsigned short wPort;
    char sUserName[64];
    char sPassword[64];
    void* cbLoginResult;
    void *pUser;
    int bUseAsynLogin;
    unsigned char byProxyType;
    unsigned char byUseUTCTime;
    unsigned char byLoginMode;
    unsigned char byHttps;
    int iProxyID;
    unsigned char byVerifyMode;
    unsigned char byRes3[119];
} NET_DVR_USER_LOGIN_INFO;

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
extern int NET_DVR_Login_V40(NET_DVR_USER_LOGIN_INFO *pLoginInfo, NET_DVR_DEVICEINFO_V40 *lpDeviceInfo);
extern int NET_DVR_Logout(int lUserID);
extern unsigned int NET_DVR_GetLastError(void);
extern int NET_DVR_GetDVRConfig(int lUserID, unsigned int dwCommand, int lChannel, void *lpOutBuf, unsigned int dwOutBufLen, unsigned int *lpBytesReturned);

typedef struct {
    char sIpV4[16];
    unsigned char byIPv6[128];
} NET_DVR_IPADDR;

typedef struct {
    NET_DVR_IPADDR struDVRIP;
    NET_DVR_IPADDR struDVRIPMask;
    unsigned int dwNetInterface;
    unsigned char byCardType;
    unsigned char byEnableDNS;
    unsigned short wMTU;
    unsigned char byMACAddr[6];
    unsigned char byEthernetPortNo;
    unsigned char bySilkScreen;
    unsigned char byUseDhcp;
    unsigned char byRes3[3];
    NET_DVR_IPADDR struGatewayIpAddr;
    NET_DVR_IPADDR struMulticastIpAddr;
    unsigned char byIPv6Address[128];
    unsigned char byIPv6AddressPrefixLen;
    unsigned char byIPv6Gateway[128];
    unsigned char byIPv6GatewayPrefixLen;
    unsigned char byIPv6Mask[128];
    unsigned char byRes[58];
} NET_DVR_ETHERNET_V30;

typedef struct {
    unsigned int dwSize;
    NET_DVR_ETHERNET_V30 struEtherNet[2];
    NET_DVR_IPADDR struRes1[2];
    NET_DVR_IPADDR struAlarmHostIpAddr;
    unsigned char byRes2[4];
    unsigned short wAlarmHostIpPort;
    unsigned char byUseDhcp;
    unsigned char byIPv6Mode;
    NET_DVR_IPADDR struDnsServer1IpAddr;
    NET_DVR_IPADDR struDnsServer2IpAddr;
    unsigned char byIpResolver[64];
    unsigned short wIpResolverPort;
    unsigned short wHttpPortNo;
    NET_DVR_IPADDR struMulticastIpAddr;
    NET_DVR_IPADDR struGatewayIpAddr;
    unsigned char byRes[128];
} NET_DVR_NETCFG_V30;

typedef struct {
    unsigned int dwDeviceStatic;
    unsigned int dwLocalDisplay;
    unsigned char byAudioChanStatus[2];
    unsigned char byRes[10];
} NET_DVR_WORKSTATE_V30;

extern int NET_DVR_GetDVRWorkState_V30(int lUserID, NET_DVR_WORKSTATE_V30 *lpWorkState);

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

func CleanupSDK() {
	mu.Lock()
	defer mu.Unlock()
	if !inited {
		return
	}
	C.NET_DVR_Cleanup()
	inited = false
}

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

func LoginV40(ip string, port uint16, username string, password string) (int32, error) {
	if !inited {
		return -1, errors.New("SDK not initialized: call InitSDK first")
	}

	cIP := C.CString(ip)
	defer C.free(unsafe.Pointer(cIP))
	cUser := C.CString(username)
	defer C.free(unsafe.Pointer(cUser))
	cPass := C.CString(password)
	defer C.free(unsafe.Pointer(cPass))

	// Try different login configurations in order of preference
	loginAttempts := []struct {
		loginMode byte
		https     byte
		desc      string
	}{
		{2, 2, "adapt mode"},  // Auto-detect both login mode and protocol
		{0, 0, "private TCP"}, // Private mode with TCP
		{1, 0, "ISAPI TCP"},   // ISAPI mode with TCP
		{0, 1, "private TLS"}, // Private mode with TLS
		{1, 1, "ISAPI TLS"},   // ISAPI mode with TLS
	}

	for _, attempt := range loginAttempts {
		var loginInfo C.NET_DVR_USER_LOGIN_INFO
		C.memset(unsafe.Pointer(&loginInfo), 0, C.sizeof_NET_DVR_USER_LOGIN_INFO)
		C.strncpy(&loginInfo.sDeviceAddress[0], cIP, C.size_t(len(ip)))
		loginInfo.wPort = C.ushort(port)
		C.strncpy(&loginInfo.sUserName[0], cUser, C.size_t(len(username)))
		C.strncpy(&loginInfo.sPassword[0], cPass, C.size_t(len(password)))
		loginInfo.byUseTransport = 0
		loginInfo.byLoginMode = C.uchar(attempt.loginMode)
		loginInfo.byHttps = C.uchar(attempt.https)

		var deviceInfo C.NET_DVR_DEVICEINFO_V40
		userID := C.NET_DVR_Login_V40(&loginInfo, &deviceInfo)
		if userID != -1 {
			return int32(userID), nil
		}
	}

	// All attempts failed, return the last error
	errCode := C.NET_DVR_GetLastError()
	return -1, fmt.Errorf("NET_DVR_Login_V40 failed with code %d (tried all login modes: adapt, private TCP, ISAPI TCP, private TLS, ISAPI TLS)", int(errCode))
}

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

type DVRWorkState struct {
	DeviceStatic   uint32
	LocalDisplay   uint32
	AudioChanState []byte
}

func KeepAlive(userID int32) error {
	if !inited {
		return errors.New("SDK not initialized: call InitSDK first")
	}
	if userID < 0 {
		return errors.New("invalid userID")
	}
	// Use a simple device config query as keepalive instead of workstate
	// This avoids V30/V40 compatibility issues
	configBuf := make([]byte, 1024)
	_, err := GetDVRConfig(userID, 100, 0, configBuf) // NET_DVR_GET_DEVICECFG
	return err
}

func GetDVRConfig(userID int32, command uint32, channel int32, out []byte) (int, error) {
	if !inited {
		return 0, errors.New("SDK not initialized: call InitSDK first")
	}
	if userID < 0 {
		return 0, errors.New("invalid userID")
	}
	if len(out) == 0 {
		return 0, errors.New("output buffer is empty")
	}
	var bytesReturned C.uint
	ret := C.NET_DVR_GetDVRConfig(C.int(userID), C.uint(command), C.int(channel), unsafe.Pointer(&out[0]), C.uint(len(out)), &bytesReturned)
	if ret == -1 {
		errCode := C.NET_DVR_GetLastError()
		return 0, fmt.Errorf("NET_DVR_GetDVRConfig failed with code %d", int(errCode))
	}
	return int(bytesReturned), nil
}

func GetDVRWorkState(userID int32) (DVRWorkState, error) {
	if !inited {
		return DVRWorkState{}, errors.New("SDK not initialized: call InitSDK first")
	}
	if userID < 0 {
		return DVRWorkState{}, errors.New("invalid userID")
	}
	var state C.NET_DVR_WORKSTATE_V30
	if C.NET_DVR_GetDVRWorkState_V30(C.int(userID), &state) == 0 {
		errCode := C.NET_DVR_GetLastError()
		// If ISAPI mode doesn't support V30 workstate, return default state instead of error
		if errCode == 189 { // NET_DVR_ISAPI_NOT_SUPPORT
			return DVRWorkState{
				DeviceStatic:   0,
				LocalDisplay:   0,
				AudioChanState: []byte{0, 0},
			}, nil
		}
		return DVRWorkState{}, fmt.Errorf("NET_DVR_GetDVRWorkState_V30 failed with code %d", int(errCode))
	}
	return DVRWorkState{
		DeviceStatic:   uint32(state.dwDeviceStatic),
		LocalDisplay:   uint32(state.dwLocalDisplay),
		AudioChanState: []byte{byte(state.byAudioChanStatus[0]), byte(state.byAudioChanStatus[1])},
	}, nil
}

func PlayBackByTime(userID int32, start, end time.Time, streamType uint8, fileIndex uint32) (int32, error) {
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
		switch errCode {
		case 17:
			return -1, fmt.Errorf("NET_DVR_PlayBackByTime_V40 failed: Wrong sequence of invoking API (error 17). This may indicate the device doesn't support playback or the login session is not properly established")
		default:
			return -1, fmt.Errorf("NET_DVR_PlayBackByTime_V40 failed with code %d", int(errCode))
		}
	}
	return int32(playHandle), nil
}

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

// NetworkInfo contains network interface information
type NetworkInfo struct {
	IP        string
	Mask      string
	Gateway   string
	MAC       string
	DHCP      bool
	Interface int
}

// GetNetworkConfig retrieves network configuration from the device
func GetNetworkConfig(userID int32) ([]NetworkInfo, error) {
	if !inited {
		return nil, errors.New("SDK not initialized: call InitSDK first")
	}
	if userID < 0 {
		return nil, errors.New("invalid userID")
	}

	configBuf := make([]byte, 1024)
	n, err := GetDVRConfig(userID, 1000, 0, configBuf) // NET_DVR_GET_NETCFG_V30
	if err != nil {
		return nil, fmt.Errorf("failed to get network config: %w", err)
	}
	if n == 0 {
		return nil, errors.New("device returned empty network config")
	}

	// Parse the NET_DVR_NETCFG_V30 structure
	var netCfg C.NET_DVR_NETCFG_V30
	if n < C.sizeof_NET_DVR_NETCFG_V30 {
		return nil, fmt.Errorf("received %d bytes, expected at least %d", n, C.sizeof_NET_DVR_NETCFG_V30)
	}

	// Copy the data into the structure
	C.memcpy(unsafe.Pointer(&netCfg), unsafe.Pointer(&configBuf[0]), C.sizeof_NET_DVR_NETCFG_V30)

	var networks []NetworkInfo
	for i := 0; i < 2; i++ { // MAX_ETHERNET = 2
		eth := netCfg.struEtherNet[i]

		// Convert MAC address to string
		mac := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
			eth.byMACAddr[0], eth.byMACAddr[1], eth.byMACAddr[2],
			eth.byMACAddr[3], eth.byMACAddr[4], eth.byMACAddr[5])

		// Convert IP addresses to strings
		ip := ipAddrToString(eth.struDVRIP)
		mask := ipAddrToString(eth.struDVRIPMask)
		gateway := ipAddrToString(eth.struGatewayIpAddr)

		networks = append(networks, NetworkInfo{
			IP:        ip,
			Mask:      mask,
			Gateway:   gateway,
			MAC:       mac,
			DHCP:      eth.byUseDhcp == 1,
			Interface: int(eth.dwNetInterface),
		})
	}

	return networks, nil
}

// ipAddrToString converts NET_DVR_IPADDR to string
func ipAddrToString(ip C.NET_DVR_IPADDR) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip.sIpV4[0], ip.sIpV4[1], ip.sIpV4[2], ip.sIpV4[3])
}
