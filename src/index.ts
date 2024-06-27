
const mimeTypeMp4Video = `video/mp4; codecs="avc1.4d4020"`
const mimeTypeMp4VideoAudio = `video/mp4; codecs="avc1.4d4020,mp4a.40.2"`

const videoElement = document.getElementById("video") as HTMLVideoElement
videoElement.autoplay = true
videoElement.muted = false

const playButton = document.getElementById("play-button") as HTMLButtonElement
playButton.onclick = () => {
    const mediaSource = new MediaSource()
    videoElement.src = URL.createObjectURL(mediaSource)

    mediaSource.onsourceopen = () => {
        const sourceBuffer = mediaSource.addSourceBuffer(mimeTypeMp4VideoAudio)
        sourceBuffer.mode = "segments"
        const sourceBuffer2 = new SourceBufferWrapper(sourceBuffer)

        const websocket = new WebSocket("ws://localhost:3000/wsmse")
        websocket.binaryType = "arraybuffer"
        websocket.onopen = () => { console.log("websocket connected") }
        websocket.onmessage = (event) => {
            sourceBuffer2.appendBuffer(new Uint8Array(event.data))
        }

        videoElement.play()
    }
}

class SourceBufferWrapper {
    sb: SourceBuffer
    buf: Uint8Array
    cap: number
    size: number
    started: boolean = false

    // use 5 Mb as default buffer capacity
    constructor(sb: SourceBuffer, cap: number = 5 * 1024 * 1024) {
        this.sb = sb
        this.cap = cap
        this.buf = new Uint8Array(cap)
        this.size = 0
        this.sb.onupdateend = this.flushBuffer
    }

    public appendBuffer(data: Uint8Array) {
        if (!data.length) { return }

        // push directly to media source buffer if we have first chunk
        if (!this.started) {
            this.sb.appendBuffer(data)
            this.started = true
            return
        }

        this.buf.set(data, this.size)
        this.size = this.size + data.length

        if (!this.sb.updating) {
            this.flushBuffer()
        }
    }

    private flushBuffer() {
        if (!this.buf || this.buf.length === 0) { return }

        // append data from buffer
        const appended = this.buf.slice(0, this.size)
        this.sb.appendBuffer(appended)

        // clear buffer
        this.buf.fill(0)
        this.size = 0
    }
}
