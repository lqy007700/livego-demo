package stream

type PacketType int

const (
	PacketVideo PacketType = iota
	PacketAudio
)

type Packet struct {
	Type      PacketType
	Timesetmp int32
	Payload   []byte
}
