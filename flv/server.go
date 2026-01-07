package flv

import (
	"livego-demo/stream"
	"log"
	"net/http"
	"strings"

	"github.com/nareix/joy4/format/flv"
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
	log.Printf("FLV play client connected: %s", streamKey)

	st := s.manager.GetOrCreate(streamKey)

	// // Add CORS headers
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Content-Type", "video/x-flv")
	// w.Header().Set("Connection", "keep-alive")
	// w.Header().Set("Transfer-Encoding", "chunked")
	// w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	// w.WriteHeader(http.StatusOK)

	// Flush headers immediately
	// if flusher, ok := w.(http.Flusher); ok {
	// 	flusher.Flush()
	// 	log.Printf("HTTP headers flushed for %s", streamKey)
	// }

	// Use joy4's FLV muxer to write FLV data
	flvMuxer := flv.NewMuxer(w)

	ch := st.Subscribe(1000)
	defer st.Unsubscribe(ch)

	log.Printf("Starting stream for %s, waiting for packets...", streamKey)
	packetCount := 0
	headerWritten := false

	for pkt := range ch {
		// Write FLV header on first packet (when stream info is available)
		if !headerWritten {
			streams := st.GetStreams()
			if streams != nil {
				if err := flvMuxer.WriteHeader(streams); err != nil {
					log.Printf("FLV muxer write header error for %s: %v", streamKey, err)
					return
				}
				log.Printf("FLV header written for %s with %d streams", streamKey, len(streams))
				headerWritten = true

				// Flush header immediately
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
					log.Printf("FLV header flushed for %s", streamKey)
				}
			} else {
				// Skip packets until stream info is available
				continue
			}
		}

		// Write the av.Packet using joy4's muxer
		if pkt.AVPacket.Data != nil {
			if err := flvMuxer.WritePacket(pkt.AVPacket); err != nil {
				log.Printf("FLV muxer write packet error for %s: %v", streamKey, err)
				return
			}

			packetCount++
			if packetCount%100 == 0 {
				log.Printf("Sent %d packets for %s", packetCount, streamKey)
			}

			// Flush data
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
	log.Printf("FLV client disconnected: %s (sent %d packets)", streamKey, packetCount)
}
