// +build go1.12

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
}

func NewPacket() *Packet {
	p := &Packet{}

	C.av_init_packet(&p.avPacket)

	p.avPacket.data = nil
	p.avPacket.size = 0

	return p
}

// Init same to NewPacket and av_init_packet
//   Initialize optional fields of a packet with default values.
//   Note, this does not touch the data and size members, which have to be
//   initialized separately.
func Init() *Packet {
	p := &Packet{}

	C.av_init_packet(&p.avPacket)

	p.avPacket.data = nil
	p.avPacket.size = 0

	return p
}

func (p *Packet) Pts() int64 {
	return int64(p.avPacket.pts)
}

func (p *Packet) SetPts(pts int64) *Packet {
	p.avPacket.pts = C.int64_t(pts)
	return p
}

func (p *Packet) Dts() int64 {
	return int64(p.avPacket.dts)
}

func (p *Packet) SetDts(val int64) *Packet {
	p.avPacket.dts = C.int64_t(val)
	return p
}

func (p *Packet) Flags() int {
	return int(p.avPacket.flags)
}

func (p *Packet) Duration() int64 {
	return int64(p.avPacket.duration)
}

func (p *Packet) SetDuration(duration int64) *Packet {
	p.avPacket.duration = C.int64_t(duration)
	return p
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

// SetData [NOT SUGGESTED] should free data later
//   p := gmf.NewPacket()
//   defer p.Free()
//   p.SetData([]byte{0x00, 0x00, 0x00, 0x01, 0x67})
//   defer p.FreeData()
func (p *Packet) SetData(data []byte) *Packet {
	p.avPacket.size = C.int(len(data))
	p.avPacket.data = (*C.uint8_t)(C.CBytes(data))
	return p
}

// FreeData free data when use SetData
//   p := gmf.NewPacket()
//   defer p.Free()
//   p.SetData([]byte{0x00, 0x00, 0x00, 0x01, 0x67})
//   defer p.FreeData()
func (p *Packet) FreeData() *Packet {
	if p.avPacket.data != nil {
		C.free(p.avPacket.data)
		p.avPacket.data = nil
		p.avPacket.size = 0
	}
	return p
}

func (p *Packet) Clone() *Packet {
	np := NewPacket()

	C.av_packet_ref(&np.avPacket, &p.avPacket)

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

func (p *Packet) Time(timebase AVRational) int {
	return int(float64(timebase.AVR().Num) / float64(timebase.AVR().Den) * float64(p.Pts()))
}
