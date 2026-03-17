package cgo

/*
#cgo CFLAGS: -I../sdk/incEn
#cgo LDFLAGS: -L../sdk/lib -lhcnetsdk

#include "HCNetSDK.h"
*/
import "C"

func Init() bool {
	ret := C.NET_DVR_Init()
	return ret != 0
}

func Login(ip string, port int, user string, pass string) int {

	cip := C.CString(ip)
	cuser := C.CString(user)
	cpass := C.CString(pass)

	id := C.NET_DVR_Login_V30(
		cip,
		C.int(port),
		cuser,
		cpass,
		nil,
	)

	return int(id)
}
