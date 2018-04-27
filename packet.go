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
	"syscall"
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

func (p *Packet) Decode(cc *CodecCtx) (*Frame, bool, error) {
	if p.frames[cc.Type()] == nil {
		p.frames[cc.Type()] = &Frame{avFrame: C.av_frame_alloc(), mediaType: cc.Type()}
	}

	return p.decode(cc, p.frames[cc.Type()])
}

func (p *Packet) decode(cc *CodecCtx, f *Frame) (*Frame, bool, error) {
	var (
		gotframe bool = false
		ret      int  = 0
	)

	if p != nil {
		ret = int(C.avcodec_send_packet(cc.avCodecCtx, &p.avPacket))
		if ret < 0 && ret != AVERROR_EOF {
			return f, false, fmt.Errorf("error sending packet - ret %d", ret)
		}
	}

	ret = int(C.avcodec_receive_frame(cc.avCodecCtx, f.avFrame))
	if ret < 0 && AvErrno(ret) != syscall.EAGAIN {
		return f, false, nil
	}

	if ret >= 0 {
		gotframe = true
	}

	return f, gotframe, nil
}

// Possible overkill, overhead of creating new frames
//
// func (p *Packet) DecodeToNewFrame(cc *CodecCtx) (*Frame, bool, error) {
// 	f := &Frame{avFrame: C.av_frame_alloc(), mediaType: cc.Type()}
// 	return p.decode(cc, f)
// }

// func (p *Packet) GetNextFrame(cc *CodecCtx) (*Frame, error) {
// 	for {
// 		if p.avPacket.size <= 0 {
// 			break
// 		}

// 		frame, ready, err := p.DecodeToNewFrame(cc)
// 		if !ready {
// 			Release(frame)

// 			if ret < 0 || err != nil {
// 				return nil, err
// 			}
// 		}

// 		C.shift_data(&p.avPacket, C.int(ret))

// 		if ready {
// 			return frame, nil
// 		}
// 	}

// 	return nil, nil
// }

func (p *Packet) Frames(cc *CodecCtx) chan *Frame {
	yield := make(chan *Frame)

	go func() {
		defer close(yield)

		for {
			frame, _, _ := p.Decode(cc)
			yield <- frame
			break

		}
	}()

	return yield
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

func (p *Packet) Duration() int {
	return int(p.avPacket.duration)
}

func (p *Packet) SetDuration(duration int) {
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
	fmt.Println(p.avPacket)
	fmt.Println("pkt:{\n", "pts:", p.avPacket.pts, "\ndts:", p.avPacket.dts, "\ndata:", string(C.GoBytes(unsafe.Pointer(p.avPacket.data), 128)), "size:", p.avPacket.size, "\n}")
}

func (p *Packet) SetStreamIndex(val int) *Packet {
	p.avPacket.stream_index = C.int(val)
	return p
}

func (p *Packet) Free() {
	C.av_packet_unref(&p.avPacket)
}
