package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/pflag"
)

var (
	address       = pflag.String("address", "localhost:3000", "")
	mediaFilePath = pflag.String("file", "./file_video.mp4", "")
)

const (
	mimeTypeMp4Video      = `video/mp4; codecs="avc1.4d4020"`
	mimeTypeMp4VideoAudio = `video/mp4; codecs="avc1.4d4020,mp4a.40.2"`
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	WriteBufferPool:  websocket.DefaultDialer.WriteBufferPool,
	CheckOrigin:      func(r *http.Request) bool { return true },
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		http.Error(w, "not websocket protocol", http.StatusBadRequest)
	},
}

func main() {
	flag.Parse()

	router := mux.NewRouter()
	router.HandleFunc("/wsmse", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		f, err := os.Open(*mediaFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		// reading file from header is necessary for media source buffer

		buffer := make([]byte, 2_000_000)
		for {
			n, err := f.Read(buffer)
			if err != nil {
				log.Fatal(err)
			}

			if err := conn.WriteMessage(websocket.BinaryMessage, buffer[:n]); err != nil {
				break
			}

			time.Sleep(time.Second)
		}
	})

	log.Println("Listen:", *address)
	if err := http.ListenAndServe(*address, router); err != nil {
		log.Fatal(err)
	}
}
