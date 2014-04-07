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

    // frame->data[idx][y * frame->linesize[idx] + x] = data;
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
	avFrame   *_Ctype_AVFrame
	mediaType int
}

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

func (this *Frame) Scale(width int, height int) *Frame {
	return nil
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

func (this *Frame) SetFormat(val int32) *Frame {
	this.avFrame.format = C.int(val) //C.int(val)
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
		C.int(this.Width()), C.int(this.Height()), C.AV_PIX_FMT_YUV420P, 32)); ret < 0 {
		return errors.New(fmt.Sprintf("Unable to allocate raw image buffer: %v", AvError(ret)))
	}

	return nil
}

func (this *Frame) SetData(idx int, lineSize int, data int) *Frame {
	C.gmf_set_frame_data(this.avFrame, C.int(idx), C.int(lineSize), (_Ctype_uint8_t)(data))

	return this
}

func (this *Frame) LineSize(idx int) int {
	return int(C.gmf_get_frame_line_size(this.avFrame, C.int(idx)))
}
