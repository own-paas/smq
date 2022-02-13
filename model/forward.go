package model

import "github.com/sestack/smq/mqtt/packets"

type Forward struct {
	ClientID string
	Packet   *packets.PublishPacket
	Timestamp int64
}
