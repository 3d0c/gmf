package gmf

/*

#cgo pkg-config: libavcodec libavutil

#include "libavcodec/avcodec.h"
#include "libavutil/frame.h"
#include "libavutil/imgutils.h"

int gmf_image_alloc(AVFrame *frame, int w, int h, int fmt, int align) {
	fprintf(stderr, "allocating, %d, %d, %d, %d\n", w, h, fmt, align);
	return av_image_alloc(frame->data, frame->linesize, w, h, fmt, align);
}

*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

type Frame struct {
	avFrame   *_Ctype_AVFrame
	mediaType int
	pts       int64
}

// In cause of this:
//  > AVFrame is typically allocated once and then reused multiple times to hold
//  > different data (e.g. a single AVFrame to hold frames received from a
//  > decoder).
// this stuff with map of singletons is used.
//
// @todo find for a better way. // see ffmpeg.c:1317
// @todo remove flush option // devmode
//
var frames map[int]*Frame = make(map[int]*Frame, 0)

func NewFrame() *Frame {
	return &Frame{avFrame: C.av_frame_alloc()}
}

func (this *Frame) Encode(cc *CodecCtx) (*Packet, bool, error) {
	var gotOutput int
	var ret int

	p := NewPacket()

	switch this.mediaType {
	case CODEC_TYPE_AUDIO:
		ret = int(C.avcodec_encode_audio2(cc.avCodecCtx, &p.avPacket, this.avFrame, (*C.int)(unsafe.Pointer(&gotOutput))))
		if ret < 0 {
			return nil, false, errors.New(fmt.Sprintf("Unable to encode video packet, averror: %s", AvError(int(ret))))
		}

	case CODEC_TYPE_VIDEO:
		cc.avCodecCtx.field_order = C.AV_FIELD_PROGRESSIVE

		ret = int(C.avcodec_encode_video2(cc.avCodecCtx, &p.avPacket, this.avFrame, (*C.int)(unsafe.Pointer(&gotOutput))))
		if ret < 0 {
			return nil, false, errors.New(fmt.Sprintf("Unable to encode video packet, averror: %s", AvError(int(ret))))
		}

	default:
		return nil, false, errors.New(fmt.Sprintf("Unknown codec type: %v", this.mediaType))
	}

	// ready := (ret == 0 && gotOutput > 0 && int(p.avPacket.size) > 0)
	ready := (gotOutput > 0)
	return p, ready, nil
}

func (this *Frame) AvPtr() unsafe.Pointer {
	return unsafe.Pointer(this.avFrame)
}

func (this *Frame) Pts() int64 {
	return int64(this.avFrame.pts)
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

func (this *Frame) SetFormat(val int32) {
	this.avFrame.format = C.int(val) //C.int(val)
}

func (this *Frame) SetWidth(val int) {
	this.avFrame.width = C.int(val)
}

func (this *Frame) SetHeight(val int) {
	this.avFrame.height = C.int(val)
}

func (this *Frame) ImgAlloc() error {
	if ret := int(C.gmf_image_alloc(this.avFrame, C.int(this.Width()), C.int(this.Height()), C.int(AV_PIX_FMT_YUV420P), 32)); ret < 0 {
		return errors.New(fmt.Sprintf("Unable to allocate raw image buffer: %v", AvError(ret)))
	}

	return nil
}
