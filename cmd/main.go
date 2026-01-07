package main

import (
	"livego-demo/flv"
	"livego-demo/rtmp"
	"livego-demo/stream"
	"log"
	"net/http"
)

func main() {
	manager := stream.NewManager()

	// http-flv
	go func() {
		http.Handle("/live/", flv.NewServer(manager))
		log.Println("HTTP-FLV server listening on :8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatal("HTTP-FLV server error:", err)
		}
	}()

	log.Println("RTMP server listening on :1935")
	rtmpServer := rtmp.NewServer(manager)
	if err := rtmpServer.Listen(":1935"); err != nil {
		log.Fatal("RTMP server error:", err)
	}
}
