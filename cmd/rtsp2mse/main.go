package main

import (
	"log"
	"net/http"
	"time"

	"github.com/deepch/vdk/av"
	mp4f "github.com/deepch/vdk/format/mp4f"
	"github.com/deepch/vdk/format/rtspv2"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/pflag"
)

var (
	rtspUrl = pflag.String("input", "rtsp://localhost:8554/test-stream", "")
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 10 * time.Second,
	// ReadBufferSize:   4096,
	// WriteBufferSize:  4096,
	WriteBufferPool: websocket.DefaultDialer.WriteBufferPool,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		http.Error(w, "not websocket protocol", http.StatusBadRequest)
	},
}

func main() {
	pflag.Parse()

	rtspClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{
		URL:              *rtspUrl,
		DisableAudio:     false,
		DialTimeout:      time.Second * 3,
		ReadWriteTimeout: 3 * time.Second,
		Debug:            false,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rtspClient.Close()

	log.Println("rtsp client connected")

	var (
		h264CodecData av.CodecData
		aacCodecData  av.CodecData
	)

	var (
		h264VideoFound = false
		aacAudioFound  = false
	)

	for _, codecData := range rtspClient.CodecData {
		switch codecData.Type().String() {
		case "H264":
			h264CodecData = codecData
			h264VideoFound = true
		case "AAC":
			aacCodecData = codecData
			aacAudioFound = true
		}
	}

	if !h264VideoFound || !aacAudioFound {
		log.Fatal("no h264 video or aac audio")
	}

	mp4Muxer := mp4f.NewMuxer(nil)

	log.Println("mp4 muxer created")

	codecs := []av.CodecData{
		h264CodecData,
		aacCodecData,
	}

	packetChan := make(chan *av.Packet, 100)
	go func() {
		for packet := range rtspClient.OutgoingPacketQueue {
			select {
			case packetChan <- packet:
			default:
			}
		}
	}()

	router := mux.NewRouter()
	router.HandleFunc("/wsmse", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		if err := mp4Muxer.WriteHeader(codecs); err != nil {
			log.Fatal(err)
		}

		// returns meta and init
		// meta - codec mime type
		// init - fmp4 header which is necessary for starting video in media source
		_, init := mp4Muxer.GetInit(codecs)

		// conn.WriteMessage(websocket.BinaryMessage, append([]byte{9}, meta...))
		conn.WriteMessage(websocket.BinaryMessage, init)

		start := false
		timeline := map[int8]time.Duration{}

		for {
			packet := <-packetChan

			if packet.IsKeyFrame {
				start = true
			}

			if !start {
				continue
			}

			timeline[packet.Idx] += packet.Duration
			packet.Time = timeline[packet.Idx]

			ready, buf, _ := mp4Muxer.WritePacket(*packet, false)
			if ready {
				if err := conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
					log.Println(err)
					return
				}
			}
		}
	})

	if err := http.ListenAndServe("localhost:3000", router); err != nil {
		log.Fatal(err)
	}
}
