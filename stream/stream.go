package stream

import (
	"log"
	"sync"

	"github.com/nareix/joy4/av"
)

type Stream struct {
	id          string
	mu          sync.RWMutex
	subscribers map[chan *Packet]struct{}
	videoHeader *Packet
	audioHeader *Packet
	metaHeader  *Packet
	closed      bool
	streams     []av.CodecData // Store stream info from RTMP
}

func NewStream(id string) *Stream {
	return &Stream{
		id:          id,
		subscribers: make(map[chan *Packet]struct{}),
	}
}

func (s *Stream) Subscribe(buffer int) chan *Packet {
	ch := make(chan *Packet, buffer)
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		close(ch)
		return nil
	}

	// Send cached headers if they exist
	if s.metaHeader != nil {
		ch <- s.metaHeader
		log.Printf("Stream %s: Sent cached metadata header to new subscriber", s.id)
	}
	if s.videoHeader != nil {
		ch <- s.videoHeader
		log.Printf("Stream %s: Sent cached video header to new subscriber", s.id)
	}
	if s.audioHeader != nil {
		ch <- s.audioHeader
		log.Printf("Stream %s: Sent cached audio header to new subscriber", s.id)
	}

	s.subscribers[ch] = struct{}{}
	return ch
}

func (s *Stream) Unsubscribe(ch chan *Packet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subscribers[ch]; ok {
		delete(s.subscribers, ch)
		close(ch)
	}
}

func (s *Stream) Broadcast(pkg *Packet) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}

	// Cache metadata header
	if pkg.Type == PacketMetadata {
		s.metaHeader = pkg
	}

	// Cache H.264 AVC sequence header (codec=7, type=0)
	if pkg.Type == PacketVideo && len(pkg.Payload) >= 2 {
		if (pkg.Payload[0]&0x0F) == 7 && pkg.Payload[1] == 0 {
			s.videoHeader = pkg
		}
	}

	// Cache AAC sequence header (format=10, type=0)
	if pkg.Type == PacketAudio && len(pkg.Payload) >= 2 {
		if (pkg.Payload[0]>>4) == 10 && pkg.Payload[1] == 0 {
			s.audioHeader = pkg
		}
	}

	for ch := range s.subscribers {
		// 不能阻塞 不能panic
		// 丢帧 不能影响整体
		select {
		case ch <- pkg:
		default:
			// 缓存满了直接丢弃
		}
	}
}

func (s *Stream) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}

	s.closed = true
	for ch := range s.subscribers {
		close(ch)
	}
	s.subscribers = nil
}

func (s *Stream) SetStreams(streams []av.CodecData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.streams = streams
	log.Printf("Stream %s: Set stream info with %d streams", s.id, len(streams))
}

func (s *Stream) GetStreams() []av.CodecData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.streams
}
