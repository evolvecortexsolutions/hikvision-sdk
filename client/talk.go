package client

import "github.com/evolvecortexsolutions/hikvision-sdk/internal/bridge"

// func (c *Client) StartTalk() (int32, error) {
// 	handle, err := bridge.StartTalkWithCallback(c.userID)
// 	if err != nil {
// 		return -1, err
// 	}
// 	return handle, nil
// }

func (c *Client) SendAudio(handle int32, data []byte) {
	bridge.SendAudio(handle, data)
}
