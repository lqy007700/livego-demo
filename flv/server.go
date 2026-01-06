package flv

import (
	"fmt"
	"livego-demo/stream"
	"net/http"
	"strings"
)

type Server struct {
	manager *stream.Manager
}

func NewServer(m *stream.Manager) *Server {
	return &Server{manager: m}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, ".flv") {
		http.NotFound(w, r)
		return
	}

	streamKey := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/live/"), ".flv")
	st := s.manager.GetOrCreate(streamKey)
	ch := st.Subscribe(100)
	defer st.Unsubscribe(ch)

	w.Header().Set("Content-Type", "video/x-flv")
	w.WriteHeader(http.StatusOK)
	fmt.Println("FLV client connected: ", streamKey)

	for pkg := range ch {
		_, err := w.Write(pkg.Payload)
		if err != nil {
			return
		}

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}
