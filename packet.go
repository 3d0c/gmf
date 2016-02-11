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
var frames map[int32]*Frame = make(map[int32]*Frame, 0)

type Packet struct {
	avPacket C.struct_AVPacket
	CgoMemoryManage
}

func NewPacket() *Packet {
	p := &Packet{}

	C.av_init_packet(&p.avPacket)

	p.avPacket.data = nil
	p.avPacket.size = 0

	return p
}

// @todo should be private
func (this *Packet) Decode(cc *CodecCtx) (*Frame, bool, int, error) {
	var gotOutput int
	var ret int = 0

	if frames[cc.Type()] == nil {
		frames[cc.Type()] = &Frame{avFrame: C.av_frame_alloc(), mediaType: cc.Type()}
	}

	switch cc.Type() {
	case AVMEDIA_TYPE_AUDIO:
		ret = int(C.avcodec_decode_audio4(cc.avCodecCtx, frames[AVMEDIA_TYPE_AUDIO].avFrame, (*C.int)(unsafe.Pointer(&gotOutput)), &this.avPacket))
		if ret < 0 {
			return nil, false, int(ret), errors.New(fmt.Sprintf("Unable to decode audio packet, averror: %s", AvError(int(ret))))
		}

		break

	case AVMEDIA_TYPE_VIDEO:
		ret = int(C.avcodec_decode_video2(cc.avCodecCtx, frames[AVMEDIA_TYPE_VIDEO].avFrame, (*C.int)(unsafe.Pointer(&gotOutput)), &this.avPacket))
		if ret < 0 {
			return nil, false, int(ret), errors.New(fmt.Sprintf("Unable to decode video packet, averror: %s", AvError(int(ret))))
		}

		break

	default:
		return nil, false, int(ret), errors.New(fmt.Sprintf("Unknown codec type: %v", cc.Type()))
	}

	return frames[cc.Type()], (gotOutput > 0), int(ret), nil
}

func (this *Packet) DecodeToNewFrame(cc *CodecCtx) (*Frame, bool, int, error) {
	f := &Frame{avFrame: C.av_frame_alloc(), mediaType: cc.Type()}
	return this.decode(cc, f)
}

func (this *Packet) decode(cc *CodecCtx, frame *Frame) (*Frame, bool, int, error) {
	var gotOutput int
	var ret int = 0

	switch cc.Type() {
	case AVMEDIA_TYPE_AUDIO:
		ret = int(C.avcodec_decode_audio4(cc.avCodecCtx, frame.avFrame, (*C.int)(unsafe.Pointer(&gotOutput)), &this.avPacket))
		if ret < 0 {
			return nil, false, int(ret), errors.New(fmt.Sprintf("Unable to decode audio packet, averror: %s", AvError(int(ret))))
		}

		break

	case AVMEDIA_TYPE_VIDEO:
		ret = int(C.avcodec_decode_video2(cc.avCodecCtx, frame.avFrame, (*C.int)(unsafe.Pointer(&gotOutput)), &this.avPacket))
		if ret < 0 {
			return nil, false, int(ret), errors.New(fmt.Sprintf("Unable to decode video packet, averror: %s", AvError(int(ret))))
		}

		break

	default:
		return nil, false, int(ret), errors.New(fmt.Sprintf("Unknown codec type: %v", cc.Type()))
	}

	return frame, (gotOutput > 0), int(ret), nil
}

func (this *Packet) GetNextFrame(cc *CodecCtx) (*Frame, error) {
	for {
		if this.avPacket.size <= 0 {
			break
		}

		frame, ready, ret, err := this.DecodeToNewFrame(cc)
		if !ready {
			Release(frame)

			if ret < 0 || err != nil {
				return nil, err
			}
		}

		C.shift_data(&this.avPacket, C.int(ret))

		if ready {
			return frame, nil
		}
	}

	return nil, nil
}

func (this *Packet) Frames(cc *CodecCtx) chan *Frame {
	yield := make(chan *Frame)

	go func() {
		defer close(yield)

		for {
			frame, ready, ret, err := this.Decode(cc)
			if ready {
				yield <- frame
			}

			if ret < 0 || err != nil {
				fmt.Println("Decoding error:", err)
				break
			}

			C.shift_data(&this.avPacket, C.int(ret))

			if this.avPacket.size <= 0 {
				break
			}
		}
	}()

	return yield
}

func (this *Packet) Pts() int64 {
	return int64(this.avPacket.pts)
}

func (this *Packet) SetPts(pts int64) {
	this.avPacket.pts = C.int64_t(pts)
}

func (this *Packet) Dts() int64 {
	return int64(this.avPacket.dts)
}

func (this *Packet) SetDts(val int64) {
	this.avPacket.dts = _Ctype_int64_t(val)
}

func (this *Packet) Flags() int {
	return int(this.avPacket.flags)
}

func (this *Packet) Duration() int {
	return int(this.avPacket.duration)
}

func (this *Packet) SetDuration(duration int) {
	this.avPacket.duration = C.int64_t(duration)
}

func (this *Packet) StreamIndex() int {
	return int(this.avPacket.stream_index)
}

func (this *Packet) Size() int {
	return int(this.avPacket.size)
}

func (this *Packet) Pos() int64 {
	return int64(this.avPacket.pos)
}

func (this *Packet) Data() []byte {
	return C.GoBytes(unsafe.Pointer(this.avPacket.data), C.int(this.avPacket.size))
}

func (this *Packet) Clone() *Packet {
	np := NewPacket()

	C.av_copy_packet(&np.avPacket, &this.avPacket)

	return np
}

func (this *Packet) Dump() {
	fmt.Println(this.avPacket)
	fmt.Println("pkt:{\n", "pts:", this.avPacket.pts, "\ndts:", this.avPacket.dts, "\ndata:", string(C.GoBytes(unsafe.Pointer(this.avPacket.data), 128)), "size:", this.avPacket.size, "\n}")
}

func (this *Packet) SetStreamIndex(val int) *Packet {
	this.avPacket.stream_index = C.int(val)
	return this
}

func (this *Packet) Free() {
	C.av_packet_unref(&this.avPacket)
}
