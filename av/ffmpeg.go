package av

import (
	"bytes"
	"fmt"
	"github.com/giorgisio/goav/avcodec"
	"github.com/giorgisio/goav/avformat"
	"github.com/giorgisio/goav/avutil"
	"github.com/giorgisio/goav/swscale"
	"image"
	"io"
	"sync"
	"time"
	"unsafe"
)

var swscalePool = SWSContextPool{oldContexts: map[int]*poolObject{}}

type Video struct {
	mutex            sync.RWMutex
	File             string
	FormatCtx        *avformat.Context
	VideoStream      *avformat.Stream
	StreamParameters *avcodec.AvCodecParameters
	Codec            *avcodec.Codec
	CodecCtx         *avcodec.Context
}

func (v *Video) Cleanup() {
	v.CodecCtx.AvcodecClose()
	v.FormatCtx.AvformatCloseInput()
}

func (v *Video) ReadFrame() (*image.RGBA, error) {

	packet := avcodec.AvPacketAlloc() // allocate a new packet
	frame := avutil.AvFrameAlloc()    // allocate a new frame

	frameRGBA := avutil.AvFrameAlloc() // allocate the RGBA frame

	v.mutex.Lock()
	defer v.mutex.Unlock()

	// calculate the size
	size := avcodec.AvpictureGetSize(avcodec.AV_PIX_FMT_RGBA, v.CodecCtx.Width(), v.CodecCtx.Height())
	buf := avutil.AvMalloc(uintptr(size))

	picture := (*avcodec.Picture)(unsafe.Pointer(frameRGBA))

	// set up the new frame with the databuffer
	picture.AvpictureFill((*uint8)(buf), avcodec.AV_PIX_FMT_RGBA, v.CodecCtx.Width(), v.CodecCtx.Height())

	swsCtx := swscalePool.GetContext(
		swscale.PixelFormat(v.CodecCtx.PixFmt()),
		v.CodecCtx.Width(),
		v.CodecCtx.Height(),
		avcodec.AV_PIX_FMT_RGBA,
		v.CodecCtx.Width(),
		v.CodecCtx.Height(),
		avcodec.SWS_BILINEAR,
	)

	var img *image.RGBA

	// read the data into the packet
	for v.FormatCtx.AvReadFrame(packet) >= 0 {
		if packet.StreamIndex() != v.VideoStream.Index() {
			continue
		}

		// send the data from the packet to the decoder
		resp := v.CodecCtx.AvcodecSendPacket(packet)
		if resp < 0 {
			return nil, avutil.ErrorFromCode(resp)
		}

		if resp >= 0 {

			resp = v.CodecCtx.AvcodecReceiveFrame((*avcodec.Frame)(unsafe.Pointer(frame)))
			//fmt.Println("received frame")
			if resp == avutil.AvErrorEOF {
				return nil, io.EOF
			} else if resp == avutil.AvErrorEAGAIN || resp == -11 {
				// we want to try again
				continue
			} else if resp < 0 {
				return nil, avutil.ErrorFromCode(resp)
			}
			if resp >= 0 {

				swscale.SwsScale2(swsCtx, avutil.Data(frame),
					avutil.Linesize(frame), 0, v.CodecCtx.Height(),
					avutil.Data(frameRGBA), avutil.Linesize(frameRGBA))

				img = toImage(frameRGBA, v.CodecCtx.Width(), v.CodecCtx.Height())
				break
			} else {
				return nil, avutil.ErrorFromCode(resp)
			}
		}
	}

	swscalePool.Return(v.CodecCtx.Width(), v.CodecCtx.Height())

	// free up
	packet.AvFreePacket()
	picture.AvpictureFree()
	avutil.AvFrameFree(frameRGBA)
	avutil.AvFrameFree(frame)

	return img, nil
}
func toImage(frame *avutil.Frame, width, height int) *image.RGBA {
	buffer := bytes.NewBuffer([]byte{})

	// Write pixel data
	for y := 0; y < height; y++ {
		data0 := avutil.Data(frame)[0]

		// pointer arithmetic
		startPos := uintptr(unsafe.Pointer(data0)) + uintptr(y)*uintptr(avutil.Linesize(frame)[0])
		buf := make([]byte, width*4)
		for i := 0; i < (width * 4); i++ {
			element := *(*uint8)(unsafe.Pointer(startPos + uintptr(i)))
			buf[i] = element
			//buf[i+3] = 255
		}

		buffer.Write(buf)
	}

	fmt.Println("expected array size =", (width*3)*height, "actual array size =", len(buffer.Bytes()))
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	img.Pix = buffer.Bytes()
	return img
}

func (v *Video) ReadFrameToImage() *image.Image {
	return nil
}

func LoadVideo(file string) (v *Video) {
	v = &Video{File: file}
	// allocate a new context
	fmtCtx := avformat.AvformatAllocContext()

	v.FormatCtx = fmtCtx

	// open the input
	avformat.AvformatOpenInput(&fmtCtx, file, nil, nil)
	fmtCtx.AvformatFindStreamInfo(nil)

	for _, stream := range fmtCtx.Streams() {
		// we only care about the video
		if stream.CodecParameters().AvCodecGetType() == avformat.AVMEDIA_TYPE_VIDEO {
			v.VideoStream = stream
			v.StreamParameters = stream.CodecParameters()
			v.Codec = avcodec.AvcodecFindDecoder(avcodec.CodecId(stream.Codec().GetCodecId()))

			// allocate the codec context
			v.CodecCtx = v.Codec.AvcodecAllocContext3()

			// copy the codec context
			v.CodecCtx.AvcodecCopyContext((*avcodec.Context)(unsafe.Pointer(stream.Codec())))

		}
	}
	v.FormatCtx.AvSeekFrameTime(v.VideoStream.Index(), 30*time.Minute, v.VideoStream.TimeBase())
	// open for reading
	v.CodecCtx.AvcodecOpen2(v.Codec, nil)

	return
}
