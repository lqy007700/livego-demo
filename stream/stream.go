package stream

import "sync"

type Stream struct {
	id          string
	mu          sync.Mutex
	subscribers map[chan *Packet]struct{}
	closed      bool
}

func NewStream(id string) *Stream {
	return &Stream{
		id:          id,
		subscribers: make(map[chan *Packet]struct{}),
	}
}

func (s *Stream) Subscribe(buffer int) chan *Packet {
	ch := make(chan *Packet, buffer)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		close(ch)
		return nil
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
