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
	"fmt"
	"unsafe"
)

type Packet struct {
	avPacket C.struct_AVPacket
	//  > AVFrame is typically allocated once and then reused multiple times to hold
	//  > different data (e.g. a single AVFrame to hold frames received from a
	//  > decoder).
	frames map[int32]*Frame

	CgoMemoryManage
}

func NewPacket() *Packet {
	p := &Packet{}

	p.initPacket()

	return p
}

func (p *Packet) initPacket() {
	C.av_init_packet(&p.avPacket)

	p.avPacket.data = nil
	p.avPacket.size = 0
	p.frames = make(map[int32]*Frame, 0)
}

func (p *Packet) Frames(cc *CodecCtx) (*Frame, error) {
	var ret int

	if p.frames[cc.Type()] == nil {
		p.frames[cc.Type()] = &Frame{avFrame: C.av_frame_alloc(), mediaType: cc.Type()}
	}

	ret = int(C.avcodec_send_packet(cc.avCodecCtx, &p.avPacket))
	if ret < 0 && ret != AVERROR_EOF {
		return nil, AvError(ret)
	}

	for {
		ret = int(C.avcodec_receive_frame(cc.avCodecCtx, p.frames[cc.Type()].avFrame))
		if ret >= 0 {
			return p.frames[cc.Type()], nil
		}
		if ret < 0 {
			break
		}
	}

	return nil, AvError(ret)
}

func (p *Packet) SendPacket(cc *CodecCtx) error {
	var ret int

	if p.frames[cc.Type()] == nil {
		p.frames[cc.Type()] = &Frame{avFrame: C.av_frame_alloc(), mediaType: cc.Type()}
	}

	ret = int(C.avcodec_send_packet(cc.avCodecCtx, &p.avPacket))
	if ret < 0 && ret != AVERROR_EOF {
		return AvError(ret)
	} else if ret < 0 {
		return AvError(ret)
	}

	return nil
}

func (p *Packet) ReceiveFrame(cc *CodecCtx) (*Frame, int) {
	var ret int

	if p.frames[cc.Type()] == nil {
		panic("frame is not initialized")
	}
	ret = int(C.avcodec_receive_frame(cc.avCodecCtx, p.frames[cc.Type()].avFrame))

	return p.frames[cc.Type()], ret
}

func ReceiveFrame(cc *CodecCtx) (*Frame, int) {
	var ret int

	frame := NewFrame()

	ret = int(C.avcodec_receive_frame(cc.avCodecCtx, frame.avFrame))

	return frame, ret
}

func (p *Packet) Pts() int64 {
	return int64(p.avPacket.pts)
}

func (p *Packet) SetPts(pts int64) {
	p.avPacket.pts = C.int64_t(pts)
}

func (p *Packet) Dts() int64 {
	return int64(p.avPacket.dts)
}

func (p *Packet) SetDts(val int64) {
	p.avPacket.dts = _Ctype_int64_t(val)
}

func (p *Packet) Flags() int {
	return int(p.avPacket.flags)
}

func (p *Packet) Duration() int64 {
	return int64(p.avPacket.duration)
}

func (p *Packet) SetDuration(duration int64) {
	p.avPacket.duration = C.int64_t(duration)
}

func (p *Packet) StreamIndex() int {
	return int(p.avPacket.stream_index)
}

func (p *Packet) Size() int {
	return int(p.avPacket.size)
}

func (p *Packet) Pos() int64 {
	return int64(p.avPacket.pos)
}

func (p *Packet) Data() []byte {
	return C.GoBytes(unsafe.Pointer(p.avPacket.data), C.int(p.avPacket.size))
}

func (p *Packet) Clone() *Packet {
	np := NewPacket()

	C.av_copy_packet(&np.avPacket, &p.avPacket)

	return np
}

func (p *Packet) Dump() {
	fmt.Printf("idx: %d\npts: %d\ndts: %d\nsize: %d\nduration:%d\npos:%d\ndata: % x\n", p.StreamIndex(), p.avPacket.pts, p.avPacket.dts, p.avPacket.size, p.avPacket.duration, p.avPacket.pos, C.GoBytes(unsafe.Pointer(p.avPacket.data), 128))
	fmt.Println("------------------------------")

}

func (p *Packet) SetStreamIndex(val int) *Packet {
	p.avPacket.stream_index = C.int(val)
	return p
}

func (p *Packet) Free() {
	C.av_packet_unref(&p.avPacket)
}

// func (p *Packet) DurationMs() int64 {
// 	return RescaleRnd(int64(p.avPacket.duration), int64(1000), int64(AV_TIME_BASE))
// }
