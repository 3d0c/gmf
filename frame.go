package gmf

/*
#cgo pkg-config: libavcodec libavutil

#include "libavcodec/avcodec.h"
#include "libavutil/frame.h"
#include "libavutil/imgutils.h"

void gmf_set_frame_data(AVFrame *frame, int idx, int l_size, uint8_t data) {
    if(!frame) {
        fprintf(stderr, "frame is NULL\n");
    }

    frame->data[idx][l_size] = data;
}

int gmf_get_frame_line_size(AVFrame *frame, int idx) {
	return frame->linesize[idx];
}

*/
import "C"

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

type Frame struct {
	avFrame   *C.struct_AVFrame
	mediaType int32
	CgoMemoryManage
}

func NewFrame() *Frame {
	return &Frame{avFrame: C.av_frame_alloc()}
}

// Deprected
// func (this *Frame) EncodeNewPacket(cc *CodecCtx) (*Packet, bool, error) {
// 	return encode(cc, this.avFrame, this.mediaType)
// }

// Deprected
// func (this *Frame) FlushNewPacket(cc *CodecCtx) (*Packet, bool, error) {
// 	return encode(cc, nil, this.mediaType)
// }

func (f *Frame) Encode(enc *CodecCtx) (*Packet, error) {
	var (
		pkt *Packet
		ret int
	)

	if pkt = NewPacket(); pkt == nil {
		return nil, fmt.Errorf("fatal error - uninitialized packet")
	}

	if ret = int(C.avcodec_send_frame(enc.avCodecCtx, f.avFrame)); ret != 0 {
		return nil, fmt.Errorf("error sending frame - %v", syscall.Errno(ret))
	}

	for {
		ret = int(C.avcodec_receive_packet(enc.avCodecCtx, &pkt.avPacket))
		if AvErrno(ret) == syscall.EAGAIN {
			break
		}
		if ret != 0 {
			return nil, fmt.Errorf("error receiving packet - %v", syscall.Errno(ret))
		}

		return pkt, nil
	}

	return nil, nil
}

func (f *Frame) Pts() int64 {
	return int64(f.avFrame.pts)
}

func (f *Frame) Unref() {
	C.av_frame_unref(f.avFrame)
}

func (f *Frame) SetPts(val int64) {
	f.avFrame.pts = (_Ctype_int64_t)(val)
}

func (f *Frame) SetBestPts() {
	f.avFrame.pts = C.av_frame_get_best_effort_timestamp(f.avFrame)
}

// AVPixelFormat for video frames, AVSampleFormat for audio
func (f *Frame) Format() int {
	return int(f.avFrame.format)
}

func (f *Frame) Width() int {
	return int(f.avFrame.width)
}

func (f *Frame) Height() int {
	return int(f.avFrame.height)
}

func (f *Frame) PktPts() int64 {
	return int64(f.avFrame.pkt_pts)
}

func (f *Frame) SetPktPts(val int64) {
	f.avFrame.pkt_pts = (_Ctype_int64_t)(val)
}

func (f *Frame) PktDts() int {
	return int(f.avFrame.pkt_dts)
}

func (f *Frame) SetPktDts(val int) {
	f.avFrame.pkt_dts = (_Ctype_int64_t)(val)
}

func (f *Frame) TimeStamp() int {
	return int(C.av_frame_get_best_effort_timestamp(f.avFrame))
}

func (f *Frame) PktPos() int {
	return int(C.av_frame_get_pkt_pos(f.avFrame))
}

func (f *Frame) PktDuration() int {
	return int(C.av_frame_get_pkt_duration(f.avFrame))
}

func (f *Frame) KeyFrame() int {
	return int(f.avFrame.key_frame)
}

func (f *Frame) NbSamples() int {
	return int(f.avFrame.nb_samples)
}

func (f *Frame) Channels() int {
	return int(f.avFrame.channels)
}

func (f *Frame) SetFormat(val int32) *Frame {
	f.avFrame.format = C.int(val)
	return f
}

func (f *Frame) SetWidth(val int) *Frame {
	f.avFrame.width = C.int(val)
	return f
}

func (f *Frame) SetHeight(val int) *Frame {
	f.avFrame.height = C.int(val)
	return f
}

func (f *Frame) ImgAlloc() error {
	if ret := int(C.av_image_alloc(
		(**C.uint8_t)(unsafe.Pointer(&f.avFrame.data)),
		(*_Ctype_int)(unsafe.Pointer(&f.avFrame.linesize)),
		C.int(f.Width()), C.int(f.Height()), int32(f.Format()), 32)); ret < 0 {
		return errors.New(fmt.Sprintf("Unable to allocate raw image buffer: %v", AvError(ret)))
	}

	return nil
}

func NewAudioFrame(sampleFormat int32, channels, nb_samples int) (*Frame, error) {
	f := NewFrame()
	f.mediaType = AVMEDIA_TYPE_AUDIO
	f.SetNbSamples(nb_samples)
	f.SetFormat(sampleFormat)
	f.SetChannelLayout(channels)

	//the codec gives us the frame size, in samples,
	//we calculate the size of the samples buffer in bytes
	size := C.av_samples_get_buffer_size(nil, C.int(channels), C.int(nb_samples),
		sampleFormat, 0)
	if size < 0 {
		return nil, errors.New("Could not get sample buffer size")
	}
	samples := (*_Ctype_uint8_t)(C.av_malloc(C.size_t(size)))
	if samples == nil {
		return nil, errors.New(fmt.Sprintf("Could not allocate %d bytes for samples buffer", size))
	}

	//setup the data pointers in the AVFrame
	ret := int(C.avcodec_fill_audio_frame(f.avFrame, C.int(channels), sampleFormat,
		samples, C.int(size), 0))
	if ret < 0 {
		return nil, errors.New("Could not setup audio frame")
	}
	return f, nil
}
func (f *Frame) SetData(idx int, lineSize int, data int) *Frame {
	C.gmf_set_frame_data(f.avFrame, C.int(idx), C.int(lineSize), (_Ctype_uint8_t)(data))

	return f
}

func (f *Frame) LineSize(idx int) int {
	return int(C.gmf_get_frame_line_size(f.avFrame, C.int(idx)))
}

func (f *Frame) CloneNewFrame() *Frame {
	return &Frame{avFrame: C.av_frame_clone(f.avFrame)}
}

func (f *Frame) Free() {
	C.av_frame_free(&f.avFrame)
}

func (f *Frame) SetNbSamples(val int) *Frame {
	f.avFrame.nb_samples = C.int(val)
	return f
}

func (f *Frame) SetChannelLayout(val int) *Frame {
	f.avFrame.channel_layout = (_Ctype_uint64_t)(val)
	return f
}

func (f *Frame) SetChannels(val int) *Frame {
	f.avFrame.channels = C.int(val)
	return f
}

func (f *Frame) SetQuality(val int) *Frame {
	f.avFrame.quality = C.int(val)
	return f
}
