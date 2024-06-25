
const codecH264 = "avc1.4d4020"
const codecAAC = "mp4a.40.2"

const mimeType = `video/mp4; codecs="` + codecH264 + "," + codecAAC + `"`
// const mimeType = `video/mp4; codecs="avc1.4d4020,mp4a.40.2"`

const videoElement = document.getElementById("video") as HTMLVideoElement
videoElement.autoplay = true
videoElement.muted = false

const playButton = document.getElementById("play-button") as HTMLButtonElement
playButton.onclick = () => {
    const mediaSource = new MediaSource()
    videoElement.src = URL.createObjectURL(mediaSource)

    mediaSource.onsourceopen = () => {
        const sourceBuffer = mediaSource.addSourceBuffer(mimeType)
        sourceBuffer.mode = "segments"
        sourceBuffer.onupdateend = () => { }

        const websocket = new WebSocket("ws://localhost:3000/wsmse")
        websocket.binaryType = "arraybuffer"
        websocket.onopen = () => { console.log("websocket connected") }
        websocket.onmessage = (event) => {
            const data = new Uint8Array(event.data)
            try {
                sourceBuffer.appendBuffer(data)
            } catch (err) { console.error(err) }
        }

        videoElement.play()
    }
}
