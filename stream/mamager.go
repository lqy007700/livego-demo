package stream

import (
	"sync"
)

type Manager struct {
	mu      sync.RWMutex
	streams map[string]*Stream
}

func NewManager() *Manager {
	return &Manager{
		streams: make(map[string]*Stream),
	}
}

func (m *Manager) GetOrCreate(id string) *Stream {
	m.mu.RLock()
	s, ok := m.streams[id]
	m.mu.RUnlock()
	if ok {
		return s
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// double check
	if s, ok := m.streams[id]; ok {
		return s
	}

	s = NewStream(id)
	m.streams[id] = s
	return s
}
