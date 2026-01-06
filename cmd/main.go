package main

import (
	"livego-demo/flv"
	"livego-demo/stream"
	"net/http"
	"time"
)

func main() {
	manager := stream.NewManager()
	// test publish stream
	go func() {
		st := manager.GetOrCreate("test")
		ticker := time.NewTicker(40 * time.Millisecond)
		defer ticker.Stop()
		for {
			<-ticker.C
			pkg := &stream.Packet{
				Type:      stream.PacketVideo,
				Timesetmp: uint32(time.Now().UnixMilli()),
				Payload:   []byte("Video-data\n"),
			}
			st.Broadcast(pkg)
		}
	}()

	http.Handle("/live/", flv.NewServer(manager))
	http.ListenAndServe(":8081", nil)
}
