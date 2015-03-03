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

func (this *Frame) EncodeNewPacket(cc *CodecCtx) (*Packet, bool, error) {
	return encode(cc, this.avFrame, this.mediaType)
}

func (this *Frame) FlushNewPacket(cc *CodecCtx) (*Packet, bool, error) {
	return encode(cc, nil, this.mediaType)
}

func encode(cc *CodecCtx, avFrame *C.struct_AVFrame, mediaType int32) (*Packet, bool, error) {
	var gotOutput int
	var ret int

	p := NewPacket()

	switch mediaType {
	case AVMEDIA_TYPE_AUDIO:
		ret = int(C.avcodec_encode_audio2(cc.avCodecCtx, &p.avPacket, avFrame, (*C.int)(unsafe.Pointer(&gotOutput))))
		if ret < 0 {
			return nil, false, errors.New(fmt.Sprintf("Unable to encode video packet, averror: %s", AvError(int(ret))))
		}

	case AVMEDIA_TYPE_VIDEO:
		cc.avCodecCtx.field_order = C.AV_FIELD_PROGRESSIVE

		ret = int(C.avcodec_encode_video2(cc.avCodecCtx, &p.avPacket, avFrame, (*C.int)(unsafe.Pointer(&gotOutput))))
		if ret < 0 {
			return nil, false, errors.New(fmt.Sprintf("Unable to encode video packet, averror: %s", AvError(int(ret))))
		}

	default:
		return nil, false, errors.New(fmt.Sprintf("Unknown codec type: %v", mediaType))
	}

	return p, (gotOutput > 0), nil
}

func (this *Frame) Pts() int {
	return int(this.avFrame.pts)
}

func (this *Frame) Unref() {
	C.av_frame_unref(this.avFrame)
}

func (this *Frame) SetPts(val int) {
	this.avFrame.pts = (_Ctype_int64_t)(val)
}

func (this *Frame) SetBestPts() {
	this.avFrame.pts = C.av_frame_get_best_effort_timestamp(this.avFrame)
}

// AVPixelFormat for video frames, AVSampleFormat for audio
func (this *Frame) Format() int {
	return int(this.avFrame.format)
}

func (this *Frame) Width() int {
	return int(this.avFrame.width)
}

func (this *Frame) Height() int {
	return int(this.avFrame.height)
}

func (this *Frame) PktPts() int {
	return int(this.avFrame.pkt_pts)
}

func (this *Frame) SetPktPts(val int) {
	this.avFrame.pkt_pts = (_Ctype_int64_t)(val)
}

func (this *Frame) PktDts() int {
	return int(this.avFrame.pkt_dts)
}

func (this *Frame) SetPktDts(val int) {
	this.avFrame.pkt_dts = (_Ctype_int64_t)(val)
}

func (this *Frame) TimeStamp() int {
	return int(C.av_frame_get_best_effort_timestamp(this.avFrame))
}

func (this *Frame) PktPos() int {
	return int(C.av_frame_get_pkt_pos(this.avFrame))
}

func (this *Frame) PktDuration() int {
	return int(C.av_frame_get_pkt_duration(this.avFrame))
}

func (this *Frame) KeyFrame() int {
	return int(this.avFrame.key_frame)
}

func (this *Frame) NbSamples() int {
	return int(this.avFrame.nb_samples)
}

func (this *Frame) Channels() int {
	return int(this.avFrame.channels)
}

func (this *Frame) SetFormat(val int32) *Frame {
	this.avFrame.format = C.int(val)
	return this
}

func (this *Frame) SetWidth(val int) *Frame {
	this.avFrame.width = C.int(val)
	return this
}

func (this *Frame) SetHeight(val int) *Frame {
	this.avFrame.height = C.int(val)
	return this
}

func (this *Frame) ImgAlloc() error {
	if ret := int(C.av_image_alloc(
		(**C.uint8_t)(unsafe.Pointer(&this.avFrame.data)),
		(*_Ctype_int)(unsafe.Pointer(&this.avFrame.linesize)),
		C.int(this.Width()), C.int(this.Height()), int32(this.Format()), 32)); ret < 0 {
		return errors.New(fmt.Sprintf("Unable to allocate raw image buffer: %v", AvError(ret)))
	}

	return nil
}

func NewAudioFrame(sampleFormat int32, channels, nb_samples int) (*Frame, error) {
	this := NewFrame()
	this.mediaType = AVMEDIA_TYPE_AUDIO
	this.SetNbSamples(nb_samples)
	this.SetFormat(sampleFormat)
	this.SetChannelLayout(channels)

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
	ret := int(C.avcodec_fill_audio_frame(this.avFrame, C.int(channels), sampleFormat,
		samples, C.int(size), 0))
	if ret < 0 {
		return nil, errors.New("Could not setup audio frame")
	}
	return this, nil
}
func (this *Frame) SetData(idx int, lineSize int, data int) *Frame {
	C.gmf_set_frame_data(this.avFrame, C.int(idx), C.int(lineSize), (_Ctype_uint8_t)(data))

	return this
}

func (this *Frame) LineSize(idx int) int {
	return int(C.gmf_get_frame_line_size(this.avFrame, C.int(idx)))
}

func (this *Frame) CloneNewFrame() *Frame {
	return &Frame{avFrame: C.av_frame_clone(this.avFrame)}
}

func (this *Frame) Free() {
	C.av_frame_free(&this.avFrame)
}

func (this *Frame) SetNbSamples(val int) *Frame {
	this.avFrame.nb_samples = C.int(val)
	return this
}

func (this *Frame) SetChannelLayout(val int) *Frame {
	this.avFrame.channel_layout = (_Ctype_uint64_t)(val)
	return this
}

func (this *Frame) SetChannels(val int) *Frame {
	this.avFrame.channels = C.int(val)
	return this
}

func (this *Frame) SetQuality(val int) *Frame {
	this.avFrame.quality = C.int(val)
	return this
}
