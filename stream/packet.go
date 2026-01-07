package stream

import "github.com/nareix/joy4/av"

type PacketType int

const (
	PacketVideo PacketType = iota
	PacketAudio
	PacketMetadata
)

type Packet struct {
	Type      PacketType
	Timestamp uint32
	Payload   []byte
	AVPacket  av.Packet // Store the original av.Packet from joy4
}
