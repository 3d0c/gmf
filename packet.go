package gmf

/*

#cgo pkg-config: libavcodec

#include "libavcodec/avcodec.h"

void shift_data(AVPacket *pkt, int offset) {
    pkt->data += offset;
    pkt->size -= offset;

    return;
}

*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// In cause of this:
//  > AVFrame is typically allocated once and then reused multiple times to hold
//  > different data (e.g. a single AVFrame to hold frames received from a
//  > decoder).
// this stuff with map of singletons is used.
//
var frames map[int]*Frame = make(map[int]*Frame, 0)

type Packet struct {
	avPacket _Ctype_AVPacket
}

func NewPacket() *Packet {
	p := &Packet{}

	C.av_init_packet(&p.avPacket)

	p.avPacket.data = nil
	p.avPacket.size = 0

	return p
}

func (this *Packet) Decode(cc *CodecCtx) (*Frame, int, error) {
	var gotOutput int

	if frames[cc.Type()] == nil {
		frames[cc.Type()] = &Frame{avFrame: C.av_frame_alloc(), mediaType: cc.Type()}
	}

	switch cc.Type() {
	case CODEC_TYPE_AUDIO:
		ret := C.avcodec_decode_audio4(cc.avCodecCtx, frames[CODEC_TYPE_AUDIO].avFrame, (*C.int)(unsafe.Pointer(&gotOutput)), &this.avPacket)
		if ret < 0 {
			return nil, 0, errors.New(fmt.Sprintf("Unable to decode audio packet, averror: %s", AvError(int(ret))))
		}

		break

	case CODEC_TYPE_VIDEO:
		// pkt->dts  = av_rescale_q(ist->dts, AV_TIME_BASE_Q, ist->st->time_base);
		// this.avPacket.dts = C.av_rescale_q(ist->dts, AV_TIME_BASE_Q, ist->st->time_base)
		ret := C.avcodec_decode_video2(cc.avCodecCtx, frames[CODEC_TYPE_VIDEO].avFrame, (*C.int)(unsafe.Pointer(&gotOutput)), &this.avPacket)
		if ret < 0 {
			return nil, 0, errors.New(fmt.Sprintf("Unable to decode video packet, averror: %s", AvError(int(ret))))
		}

		break

	default:
		return nil, 0, errors.New(fmt.Sprintf("Unknown codec type: %v", cc.Type()))
	}

	return frames[cc.Type()], gotOutput, nil
}

func (this *Packet) DecodeV2(cc *CodecCtx) *Frame {
	var gotOutput int

	if frames[cc.Type()] == nil {
		frames[cc.Type()] = &Frame{avFrame: C.av_frame_alloc(), mediaType: cc.Type()}
	}

	ret := C.avcodec_decode_video2(cc.avCodecCtx, frames[CODEC_TYPE_VIDEO].avFrame, (*C.int)(unsafe.Pointer(&gotOutput)), &this.avPacket)
	if ret < 0 {
		return nil
		fmt.Printf("Unable to decode video packet, averror: %s", AvError(int(ret)))
	}

	if gotOutput != 0 {
		frames[cc.Type()].avFrame.pts = C.av_frame_get_best_effort_timestamp(frames[cc.Type()].avFrame)

		return frames[cc.Type()]
	}

	return nil
}

func (this *Packet) DecodeV(cc *CodecCtx) chan *Frame {
	var gotOutput int

	if frames[cc.Type()] == nil {
		frames[cc.Type()] = &Frame{avFrame: C.av_frame_alloc(), mediaType: cc.Type()}
	}

	yield := make(chan *Frame)

	go func() {
		for {
			ret := C.avcodec_decode_video2(cc.avCodecCtx, frames[CODEC_TYPE_VIDEO].avFrame, (*C.int)(unsafe.Pointer(&gotOutput)), &this.avPacket)
			if ret < 0 {
				break
				fmt.Printf("Unable to decode video packet, averror: %s", AvError(int(ret)))
			}

			if ret == 0 && gotOutput == 0 {
				break
			}

			if ret == 0 {
				continue
			}

			if gotOutput != 0 {
				// frame->pts = av_frame_get_best_effort_timestamp(frame);
				frames[cc.Type()].avFrame.pts = C.av_frame_get_best_effort_timestamp(frames[cc.Type()].avFrame)
				yield <- frames[cc.Type()]
			}

			C.shift_data(&this.avPacket, C.int(ret))

			if this.avPacket.size <= 0 {
				break
			}
		}

		close(yield)
	}()

	return yield
}

func (this *Packet) Pts() int {
	return int(this.avPacket.pts)
}

func (this *Packet) SetPts(pts int) {
	this.avPacket.pts = C.int64_t(pts)
}

func (this *Packet) Dts() int {
	return int(this.avPacket.dts)
}

func (this *Packet) SetDts(val int) {
	this.avPacket.dts = _Ctype_int64_t(val)
}

func (this *Packet) Duration() int {
	return int(this.avPacket.duration)
}

func (this *Packet) SetDuration(duration int) {
	this.avPacket.duration = C.int(duration)
}

func (this *Packet) StreamIndex() int {
	return int(this.avPacket.stream_index)
}

func (this *Packet) Size() int {
	return int(this.avPacket.size)
}

func (this *Packet) Data() []byte {
	return C.GoBytes(unsafe.Pointer(this.avPacket.data), C.int(this.avPacket.size))
}

func (this *Packet) Dump() {
	fmt.Println("pkt:{\n", "pts:", this.avPacket.pts, "\ndts:", this.avPacket.dts, "\ndata:", string(C.GoBytes(unsafe.Pointer(this.avPacket.data), 128)), "size:", this.avPacket.size, "\n}")
}
