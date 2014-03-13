package gmf

/*

#cgo pkg-config: libavcodec

#include "libavutil/channel_layout.h"
#include "libavutil/samplefmt.h"
#include "libavutil/opt.h"

*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
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

	// Temp stuff, only for testing
	if codec.Type() == CODEC_TYPE_AUDIO {
		codecctx.bit_rate = 64000
		codecctx.sample_rate = 44100
		codecctx.sample_fmt = C.AV_SAMPLE_FMT_S16

		codecctx.channel_layout = C.AV_CH_LAYOUT_STEREO // select_channel_layout(codec);
		codecctx.channels = C.av_get_channel_layout_nb_channels(C.AV_CH_LAYOUT_STEREO)
	}

	if codec.Type() == CODEC_TYPE_VIDEO {
		codecctx.bit_rate = 400000
		codecctx.width = 426
		codecctx.height = 240
		codecctx.pix_fmt = AV_PIX_FMT_YUV420P
		codecctx.time_base = C.AVRational{1, 25}
		codecctx.gop_size = 12
		codecctx.max_b_frames = 1
	}
	// eof Temp stuff

	result.avCodecCtx = codecctx

	return result
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
