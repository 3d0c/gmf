package gmf

/*

#cgo pkg-config: libavcodec libavutil

#include "libavcodec/avcodec.h"
#include "libavutils/frame.h"

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
// @todo find for a better way.
// @todo use something more DRY instead of switch
//
var frames map[int]*Frame = make(map[int]*Frame, 0)

func (this *Frame) Encode(cc *CodecCtx) (*Packet, bool, error) {
	var gotOutput int
	var ret int

	p := NewPacket()
	// p.avPacket.pts = C.AV_NOPTS_VALUE
	// p.avPacket.dts = C.AV_NOPTS_VALUE

	switch this.mediaType {
	case CODEC_TYPE_AUDIO:
		ret = int(C.avcodec_encode_audio2(cc.avCodecCtx, &p.avPacket, this.avFrame, (*C.int)(unsafe.Pointer(&gotOutput))))
		if ret < 0 {
			return nil, false, errors.New(fmt.Sprintf("Unable to encode video packet, averror: %s", AvError(int(ret))))
		}

	case CODEC_TYPE_VIDEO:
		ret = int(C.avcodec_encode_video2(cc.avCodecCtx, &p.avPacket, this.avFrame, (*C.int)(unsafe.Pointer(&gotOutput))))
		if ret < 0 {
			return nil, false, errors.New(fmt.Sprintf("Unable to encode video packet, averror: %s", AvError(int(ret))))
		}

	default:
		return nil, false, errors.New(fmt.Sprintf("Unknown codec type: %v", this.mediaType))
	}

	ready := (ret == 0 && gotOutput > 0 && int(p.avPacket.size) > 0)
	return p, ready, nil
}

func (this *Frame) Pts() int64 {
	return int64(this.avFrame.pts)
}

func (this *Frame) Unref() {
	C.av_frame_unref(this.avFrame)
}

func (this *Frame) SetPts(pts int) int64 {
	if this.avFrame == nil {
		return 0
	}

	// this.avFrame.pts = this.avFrame.pts + (_Ctype_int64_t)(int64(pts))
	this.pts += int64(pts)
	// fmt.Println(this.in, "SetPts:", this.avFrame.pts, this.pts)
	this.avFrame.pts = (_Ctype_int64_t)(this.pts)
	return int64(this.avFrame.pts)
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

func (this *Frame) TimeStamp() int {
	return int(C.av_frame_get_best_effort_timestamp(this.avFrame))
}

func (this *Frame) PktPos() int {
	return int(C.av_frame_get_pkt_pos(this.avFrame))
}

func (this *Frame) PktDuration() int {
	return int(C.av_frame_get_pkt_duration(this.avFrame))
}
