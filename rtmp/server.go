package rtmp

import (
	"livego-demo/stream"
	"log"
	"strings"
	"time"

	"github.com/nareix/joy4/format/rtmp"
)

type Server struct {
	mamager *stream.Manager
}

func NewServer(m *stream.Manager) *Server {
	return &Server{mamager: m}
}

func (s *Server) Listen(addr string) error {
	rtmp.Debug = false
	srv := rtmp.Server{
		Addr:          addr,
		HandlePublish: s.HandlerPush,
	}

	return srv.ListenAndServe()
}

// RTMP push handler
func (s *Server) HandlerPush(conn *rtmp.Conn) {
	streamKey := strings.TrimPrefix(conn.URL.Path, "/live/")
	log.Printf("RTMP push client connected: %s (path: %s)", streamKey, conn.URL.Path)

	st := s.mamager.GetOrCreate(streamKey)

	// Get stream info
	streams, err := conn.Streams()
	if err != nil {
		log.Printf("RTMP get streams error for %s: %v", streamKey, err)
		return
	}

	// Store stream info
	st.SetStreams(streams)

	log.Printf("RTMP stream info for %s: %d streams", streamKey, len(streams))
	packetCount := 0

	for {
		pkt, err := conn.ReadPacket()
		if err != nil {
			log.Printf("RTMP push disconnected for %s: %v (sent %d packets)", streamKey, err, packetCount)
			return
		}

		packetCount++
		// if packetCount == 1 || packetCount%500 == 0 {
		// 	log.Printf("RTMP %s: received %d packets", streamKey, packetCount)
		// }

		// Determine packet type
		var pType stream.PacketType
		if int(pkt.Idx) < len(streams) {
			codec := streams[pkt.Idx]
			if codec.Type().IsVideo() {
				pType = stream.PacketVideo
			} else if codec.Type().IsAudio() {
				pType = stream.PacketAudio
			} else {
				pType = stream.PacketMetadata
			}
		} else {
			continue
		}

		// Store the av.Packet
		p := &stream.Packet{
			Type:      pType,
			Timestamp: uint32(pkt.Time / time.Millisecond),
			Payload:   pkt.Data,
			AVPacket:  pkt, // Store the original av.Packet
		}
		st.Broadcast(p)
	}
}
