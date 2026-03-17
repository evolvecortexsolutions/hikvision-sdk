package login

import "github.com/evolvecortexsolutions/hikvision-sdk/cgo"

func HikLogin(ip string, port int, username string, password string) int {

	cgo.Init()

	userID := cgo.Login(
		ip,
		port,
		username,
		password,
	)

	return userID
}
