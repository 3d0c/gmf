package gmf

/*

#cgo pkg-config: libavcodec libavutil

#include <string.h>

#include "libavcodec/avcodec.h"
#include "libavutil/channel_layout.h"
#include "libavutil/samplefmt.h"
#include "libavutil/opt.h"
#include "libavutil/mem.h"

*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

var (
	AV_CODEC_ID_MPEG1VIDEO   int = C.AV_CODEC_ID_MPEG1VIDEO
	AV_CODEC_ID_H264         int = C.AV_CODEC_ID_H264
	CODEC_FLAG_GLOBAL_HEADER int = C.CODEC_FLAG_GLOBAL_HEADER
)

type CodecCtx struct {
	codec      *Codec
	avCodecCtx *_Ctype_AVCodecContext
}

func NewCodecCtx(codec *Codec) *CodecCtx {
	result := &CodecCtx{codec: codec}

	codecctx := C.avcodec_alloc_context3(codec.avCodec)
	if codecctx == nil {
		return nil
	}

	C.avcodec_get_context_defaults3(codecctx, codec.avCodec)

	result.avCodecCtx = codecctx

	return result
}

func (this *CodecCtx) CopyCtx(ist *Stream) {
	codec := this.avCodecCtx
	icodec := ist.GetCodecCtx().avCodecCtx

	codec.bits_per_raw_sample = icodec.bits_per_raw_sample
	codec.chroma_sample_location = icodec.chroma_sample_location

	codec.codec_id = icodec.codec_id
	codec.codec_type = icodec.codec_type

	codec.codec_tag = icodec.codec_tag

	codec.bit_rate = icodec.bit_rate
	codec.rc_max_rate = icodec.rc_max_rate

	codec.rc_buffer_size = icodec.rc_buffer_size

	codec.field_order = icodec.field_order

	codec.extradata = (*_Ctype_uint8_t)(C.av_mallocz((_Ctype_size_t)((C.uint64_t)(icodec.extradata_size) + C.FF_INPUT_BUFFER_PADDING_SIZE)))

	C.memcpy(unsafe.Pointer(codec.extradata), unsafe.Pointer(icodec.extradata), (_Ctype_size_t)(icodec.extradata_size))
	codec.extradata_size = icodec.extradata_size
	codec.bits_per_coded_sample = icodec.bits_per_coded_sample

	// fmt.Println("ist.avStream.time_base", ist.avStream.time_base)
	// codec.time_base = ist.avStream.time_base

	codec.pix_fmt = icodec.pix_fmt
	codec.width = icodec.width
	codec.height = icodec.height
	codec.has_b_frames = icodec.has_b_frames

	// av_reduce(&codec->time_base.num, &codec->time_base.den, codec->time_base.num, codec->time_base.den, INT_MAX);

	// C.av_reduce(codec.time_base.num, codec.time_base.den, codec.time_base.num, )
	codec.time_base = icodec.time_base
	codec.time_base.num *= icodec.ticks_per_frame

	fmt.Println("codec.time_base:", codec.time_base)
}

func (this *CodecCtx) Open(opts *Options) error {
	if averr := C.avcodec_open2(this.avCodecCtx, this.codec.avCodec, nil); averr < 0 {
		return errors.New(fmt.Sprintf("Error opening codec '%s:%s', averror: %s", this.codec.Name(), this.codec.LongName(), AvError(int(averr))))
	}

	return nil
}

func (this *CodecCtx) SetOpt() {
	// mock
	C.av_opt_set_int(unsafe.Pointer(this.avCodecCtx), C.CString("refcounted_frames"), 1, 0)
}

func (this *CodecCtx) Id() int {
	return int(this.avCodecCtx.codec_id)
}

func (this *CodecCtx) Type() int {
	return int(this.avCodecCtx.codec_type)
}

func (this *CodecCtx) Width() int {
	return int(this.avCodecCtx.width)
}

func (this *CodecCtx) Height() int {
	return int(this.avCodecCtx.height)
}

func (this *CodecCtx) PixFmt() int32 {
	return int32(this.avCodecCtx.pix_fmt)
}

func (this *CodecCtx) GetProfile() int {
	return int(this.avCodecCtx.profile)
}

func (this *CodecCtx) SetProfile(profile int) {
	this.avCodecCtx.profile = C.int(profile)
}

func (this *CodecCtx) TimeBase() AVRational {
	return AVRational(this.avCodecCtx.time_base)
}

func (this *CodecCtx) SetBitRate(val int) {
	this.avCodecCtx.bit_rate = C.int(val)
}

func (this *CodecCtx) SetWidth(val int) {
	this.avCodecCtx.width = C.int(val)
}

func (this *CodecCtx) SetHeight(val int) {
	this.avCodecCtx.height = C.int(val)
}

func (this *CodecCtx) SetTimeBase(val AVR) {
	this.avCodecCtx.time_base.num = C.int(val.Num)
	this.avCodecCtx.time_base.den = C.int(val.Den)
}

func (this *CodecCtx) SetGopSize(val int) {
	this.avCodecCtx.gop_size = C.int(val)
}

func (this *CodecCtx) SetMaxBFrames(val int) {
	this.avCodecCtx.max_b_frames = C.int(val)
}

func (this *CodecCtx) SetPixFmt(val int32) {
	this.avCodecCtx.pix_fmt = val
}

func (this *CodecCtx) SetFlag(flag int) {
	this.avCodecCtx.flags |= C.int(flag)
}
