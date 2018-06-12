package gmf

/*

#cgo pkg-config: libavcodec libavutil

#include <string.h>

#include "libavcodec/avcodec.h"
#include "libavutil/channel_layout.h"
#include "libavutil/samplefmt.h"
#include "libavutil/opt.h"
#include "libavutil/mem.h"

static int check_sample_fmt(AVCodec *codec, enum AVSampleFormat sample_fmt) {
    const enum AVSampleFormat *p = codec->sample_fmts;

    while (*p != AV_SAMPLE_FMT_NONE) {
        if (*p == sample_fmt)
            return 1;
        p++;
    }
    return 0;
}

static int select_sample_rate(AVCodec *codec) {
    const int *p;
    int best_samplerate = 0;

    if (!codec->supported_samplerates)
        return 44100;

    p = codec->supported_samplerates;
    while (*p) {
        best_samplerate = FFMAX(*p, best_samplerate);
        p++;
    }
    return best_samplerate;
}

static int select_channel_layout(AVCodec *codec) {
    const uint64_t *p;
    uint64_t best_ch_layout = 0;
    int best_nb_channels    = 0;

    if (!codec->channel_layouts)
        return AV_CH_LAYOUT_STEREO;

    p = codec->channel_layouts;
    while (*p) {
        int nb_channels = av_get_channel_layout_nb_channels(*p);

        if (nb_channels > best_nb_channels) {
            best_ch_layout   = *p;
            best_nb_channels = nb_channels;
        }
        p++;
    }
    return best_ch_layout;
}

static void call_av_freep(AVCodecContext *out){
    return av_freep(&out);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
	//	"log"
)

var (
	AV_CODEC_ID_MPEG1VIDEO int = C.AV_CODEC_ID_MPEG1VIDEO
	AV_CODEC_ID_MPEG2VIDEO int = C.AV_CODEC_ID_MPEG2VIDEO
	AV_CODEC_ID_H264       int = C.AV_CODEC_ID_H264
	AV_CODEC_ID_MPEG4      int = C.AV_CODEC_ID_MPEG4
	AV_CODEC_ID_JPEG2000   int = C.AV_CODEC_ID_JPEG2000
	AV_CODEC_ID_MJPEG      int = C.AV_CODEC_ID_MJPEG
	AV_CODEC_ID_MSMPEG4V1  int = C.AV_CODEC_ID_MSMPEG4V1
	AV_CODEC_ID_MSMPEG4V2  int = C.AV_CODEC_ID_MSMPEG4V2
	AV_CODEC_ID_MSMPEG4V3  int = C.AV_CODEC_ID_MSMPEG4V3
	AV_CODEC_ID_WMV1       int = C.AV_CODEC_ID_WMV1
	AV_CODEC_ID_WMV2       int = C.AV_CODEC_ID_WMV2
	AV_CODEC_ID_FLV1       int = C.AV_CODEC_ID_FLV1
	AV_CODEC_ID_PNG        int = C.AV_CODEC_ID_PNG
	AV_CODEC_ID_TIFF       int = C.AV_CODEC_ID_TIFF
	AV_CODEC_ID_GIF        int = C.AV_CODEC_ID_GIF
	AV_CODEC_ID_RAWVIDEO   int = C.AV_CODEC_ID_RAWVIDEO

	CODEC_FLAG_GLOBAL_HEADER int = C.AV_CODEC_FLAG_GLOBAL_HEADER
	FF_MB_DECISION_SIMPLE    int = C.FF_MB_DECISION_SIMPLE
	FF_MB_DECISION_BITS      int = C.FF_MB_DECISION_BITS
	FF_MB_DECISION_RD        int = C.FF_MB_DECISION_RD

	AV_SAMPLE_FMT_U8  int32 = C.AV_SAMPLE_FMT_U8
	AV_SAMPLE_FMT_S16 int32 = C.AV_SAMPLE_FMT_S16
	AV_SAMPLE_FMT_S32 int32 = C.AV_SAMPLE_FMT_S32
	AV_SAMPLE_FMT_FLT int32 = C.AV_SAMPLE_FMT_FLT
	AV_SAMPLE_FMT_DBL int32 = C.AV_SAMPLE_FMT_DBL

	AV_SAMPLE_FMT_U8P  int32 = C.AV_SAMPLE_FMT_U8P
	AV_SAMPLE_FMT_S16P int32 = C.AV_SAMPLE_FMT_S16P
	AV_SAMPLE_FMT_S32P int32 = C.AV_SAMPLE_FMT_S32P
	AV_SAMPLE_FMT_FLTP int32 = C.AV_SAMPLE_FMT_FLTP
	AV_SAMPLE_FMT_DBLP int32 = C.AV_SAMPLE_FMT_DBLP
)

type SampleFmt int

type CodecCtx struct {
	codec      *Codec
	avCodecCtx *C.struct_AVCodecContext
	CgoMemoryManage
}

func NewCodecCtx(codec *Codec, options ...[]*Option) *CodecCtx {
	result := &CodecCtx{codec: codec}

	codecctx := C.avcodec_alloc_context3(codec.avCodec)
	if codecctx == nil {
		return nil
	}

	C.avcodec_get_context_defaults3(codecctx, codec.avCodec)

	result.avCodecCtx = codecctx

	// we're really expecting only one options-array —
	// variadic arg is used for backward compatibility
	if len(options) == 1 {
		for _, option := range options[0] {
			option.Set(result.avCodecCtx)
		}
	}

	return result
}

func (this *CodecCtx) CopyExtra(ist *Stream) *CodecCtx {
	codec := this.avCodecCtx
	icodec := ist.CodecCtx().avCodecCtx

	codec.bits_per_raw_sample = icodec.bits_per_raw_sample
	codec.chroma_sample_location = icodec.chroma_sample_location

	codec.codec_id = icodec.codec_id
	codec.codec_type = icodec.codec_type

	// codec.codec_tag = icodec.codec_tag

	codec.rc_max_rate = icodec.rc_max_rate
	codec.rc_buffer_size = icodec.rc_buffer_size

	codec.field_order = icodec.field_order

	codec.extradata = (*_Ctype_uint8_t)(C.av_mallocz((_Ctype_size_t)((C.uint64_t)(icodec.extradata_size) + C.AV_INPUT_BUFFER_PADDING_SIZE)))

	C.memcpy(unsafe.Pointer(codec.extradata), unsafe.Pointer(icodec.extradata), (_Ctype_size_t)(icodec.extradata_size))
	codec.extradata_size = icodec.extradata_size
	codec.bits_per_coded_sample = icodec.bits_per_coded_sample

	codec.has_b_frames = icodec.has_b_frames

	return this
}

func (this *CodecCtx) CopyBasic(ist *Stream) *CodecCtx {
	codec := this.avCodecCtx
	icodec := ist.CodecCtx().avCodecCtx

	codec.bit_rate = icodec.bit_rate
	codec.pix_fmt = icodec.pix_fmt
	codec.width = icodec.width
	codec.height = icodec.height

	codec.time_base = icodec.time_base
	codec.time_base.num *= icodec.ticks_per_frame

	codec.sample_fmt = icodec.sample_fmt
	codec.sample_rate = icodec.sample_rate
	codec.channels = icodec.channels

	codec.channel_layout = icodec.channel_layout

	return this
}

func (this *CodecCtx) Open(dict *Dict) error {
	if this.IsOpen() {
		return nil
	}

	var avDict *C.struct_AVDictionary
	if dict != nil {
		avDict = dict.avDict
	}

	if averr := C.avcodec_open2(this.avCodecCtx, this.codec.avCodec, &avDict); averr < 0 {
		return errors.New(fmt.Sprintf("Error opening codec '%s:%s', averror: %s", this.codec.Name(), this.codec.LongName(), AvError(int(averr))))
	}

	return nil
}

func (this *CodecCtx) Close() {
	if nil != this.avCodecCtx {
		C.avcodec_close(this.avCodecCtx)
		this.avCodecCtx = nil
	}
}

func (this *CodecCtx) Free() {
	this.CloseAndRelease()
}

func (this *CodecCtx) CloseAndRelease() {
	this.Close()
	C.call_av_freep(this.avCodecCtx)
}

// @todo
func (this *CodecCtx) SetOpt() {
	// mock
	C.av_opt_set_int(unsafe.Pointer(this.avCodecCtx), C.CString("refcounted_frames"), 1, 0)
}

func (this *CodecCtx) Codec() *Codec {
	return &Codec{avCodec: this.avCodecCtx.codec}
}

func (this *CodecCtx) Id() int {
	return int(this.avCodecCtx.codec_id)
}

func (this *CodecCtx) Type() int32 {
	return int32(this.avCodecCtx.codec_type)
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

func (this *CodecCtx) FrameSize() int {
	return int(this.avCodecCtx.frame_size)
}

func (this *CodecCtx) SampleFmt() int32 {
	return this.avCodecCtx.sample_fmt
}

func (this *CodecCtx) SampleRate() int {
	return int(this.avCodecCtx.sample_rate)
}

func (this *CodecCtx) Profile() int {
	return int(this.avCodecCtx.profile)
}

func (this *CodecCtx) IsOpen() bool {
	return (int(C.avcodec_is_open(this.avCodecCtx)) > 0)
}

func (this *CodecCtx) SetProfile(profile int) *CodecCtx {
	this.avCodecCtx.profile = C.int(profile)
	return this
}

func (this *CodecCtx) TimeBase() AVRational {
	return AVRational(this.avCodecCtx.time_base)
}

func (this *CodecCtx) ChannelLayout() int {
	return int(this.avCodecCtx.channel_layout)
}
func (this *CodecCtx) SetChannelLayout(channelLayout int) {
	this.avCodecCtx.channel_layout = C.uint64_t(channelLayout)
}

func (this *CodecCtx) BitRate() int {
	return int(this.avCodecCtx.bit_rate)
}

func (this *CodecCtx) Channels() int {
	return int(this.avCodecCtx.channels)
}

func (this *CodecCtx) SetBitRate(val int) *CodecCtx {
	this.avCodecCtx.bit_rate = C.int64_t(val)
	return this
}

func (this *CodecCtx) SetWidth(val int) *CodecCtx {
	this.avCodecCtx.width = C.int(val)
	return this
}

func (this *CodecCtx) SetHeight(val int) *CodecCtx {
	this.avCodecCtx.height = C.int(val)
	return this
}

func (this *CodecCtx) SetDimension(w, h int) *CodecCtx {
	this.avCodecCtx.width = C.int(w)
	this.avCodecCtx.height = C.int(h)
	return this
}

func (this *CodecCtx) SetTimeBase(val AVR) *CodecCtx {
	this.avCodecCtx.time_base.num = C.int(val.Num)
	this.avCodecCtx.time_base.den = C.int(val.Den)
	return this
}

func (this *CodecCtx) SetGopSize(val int) *CodecCtx {
	this.avCodecCtx.gop_size = C.int(val)
	return this
}

func (this *CodecCtx) SetMaxBFrames(val int) *CodecCtx {
	this.avCodecCtx.max_b_frames = C.int(val)
	return this
}

func (this *CodecCtx) SetPixFmt(val int32) *CodecCtx {
	this.avCodecCtx.pix_fmt = val
	return this
}

func (this *CodecCtx) SetFlag(flag int) *CodecCtx {
	this.avCodecCtx.flags |= C.int(flag)
	return this
}

func (this *CodecCtx) SetMbDecision(val int) *CodecCtx {
	this.avCodecCtx.mb_decision = C.int(val)
	return this
}

func (this *CodecCtx) SetSampleFmt(val int32) *CodecCtx {
	if int(C.check_sample_fmt(this.codec.avCodec, val)) == 0 {
		panic(fmt.Sprintf("encoder doesn't support sample format %s", GetSampleFmtName(val)))
	}

	this.avCodecCtx.sample_fmt = val
	return this
}

func (this *CodecCtx) SetSampleRate(val int) *CodecCtx {
	this.avCodecCtx.sample_rate = C.int(val)
	return this
}

var (
	FF_COMPLIANCE_VERY_STRICT  int = C.FF_COMPLIANCE_VERY_STRICT
	FF_COMPLIANCE_STRICT       int = C.FF_COMPLIANCE_STRICT
	FF_COMPLIANCE_NORMAL       int = C.FF_COMPLIANCE_NORMAL
	FF_COMPLIANCE_UNOFFICIAL   int = C.FF_COMPLIANCE_UNOFFICIAL
	FF_COMPLIANCE_EXPERIMENTAL int = C.FF_COMPLIANCE_EXPERIMENTAL
)

func (this *CodecCtx) SetStrictCompliance(val int) *CodecCtx {
	this.avCodecCtx.strict_std_compliance = C.int(val)
	return this
}

func (this *CodecCtx) SetHasBframes(val int) *CodecCtx {
	this.avCodecCtx.has_b_frames = C.int(val)
	return this
}

func (this *CodecCtx) SetChannels(val int) *CodecCtx {
	this.avCodecCtx.channels = C.int(val)
	return this
}

func (this *CodecCtx) SelectSampleRate() int {
	return int(C.select_sample_rate(this.codec.avCodec))
}

func (this *CodecCtx) SelectChannelLayout() int {
	return int(C.select_channel_layout(this.codec.avCodec))
}

func (this *CodecCtx) FlushBuffers() {
	C.avcodec_flush_buffers(this.avCodecCtx)
}

func (this *CodecCtx) Dump() {
	fmt.Println(this.avCodecCtx)
}
