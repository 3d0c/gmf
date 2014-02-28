package gmf

/*
#cgo pkg-config: libavcodec libavutil

#include "libavcodec/avcodec.h"

#include "libavutil/samplefmt.h"
#include "libavutil/channel_layout.h"
#include "libavutil/pixfmt.h"

*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

var (
	CODEC_TYPE_AUDIO int = C.AVMEDIA_TYPE_AUDIO
	CODEC_TYPE_VIDEO int = C.AVMEDIA_TYPE_VIDEO
)

func init() {
	C.avcodec_register_all()
}

// AVCodec
//
type Codec struct {
	avCodec *_Ctype_AVCodec
}

func NewDecoder(i interface{}) (*Codec, error) {
	var avc *_Ctype_AVCodec

	switch t := i.(type) {
	case string:
		cname := C.CString(i.(string))
		defer C.free(unsafe.Pointer(cname))

		avc = C.avcodec_find_decoder_by_name(cname)
		break

	case int:
		avc = C.avcodec_find_decoder(uint32(i.(int)))
		break

	default:
		return nil, errors.New(fmt.Sprintf("Unable to find codec, unexpected arguments type '%v'", t))
	}

	if avc == nil {
		return nil, errors.New(fmt.Sprintf("Unable to find codec by value '%v'", i))
	}

	return &Codec{avCodec: avc}, nil
}

func NewEncoder(i interface{}) (*Codec, error) {
	var avc *_Ctype_AVCodec

	switch t := i.(type) {
	case string:
		cname := C.CString(i.(string))
		defer C.free(unsafe.Pointer(cname))

		avc = C.avcodec_find_encoder_by_name(cname)
		break

	case int:
		avc = C.avcodec_find_encoder(uint32(i.(int)))
		break

	default:
		return nil, errors.New(fmt.Sprintf("Unable to find codec, unexpected arguments type '%v'", t))
	}

	if avc == nil {
		return nil, errors.New(fmt.Sprintf("Unable to find codec by value '%v'", i))
	}

	return &Codec{avCodec: avc}, nil
}

func (this *Codec) Name() string {
	return C.GoString(this.avCodec.name)
}

func (this *Codec) LongName() string {
	return C.GoString(this.avCodec.long_name)
}

// > ...field names that are keywords in Go can be
// > accessed by prefixing them with an underscore
func (this *Codec) Type() int {
	return int(this.avCodec._type)
}

// Options
//
type Options struct{}

// AVCodecContext
//
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
		codecctx.pix_fmt = C.AV_PIX_FMT_YUV420P
		codecctx.time_base = C.AVRational{1, 25}
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

func (this *CodecCtx) Id() int {
	return int(this.avCodecCtx.codec_id)
}

func (this *CodecCtx) Type() int {
	return int(this.avCodecCtx.codec_type)
}

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

func (this *Packet) Pts() int {
	return int(this.avPacket.pts)
}

func (this *Packet) StreamIndex() int {
	return int(this.avPacket.stream_index)
}

func (this *Packet) Size() int {
	return int(this.avPacket.size)
}

type Frame struct {
	avFrame   *_Ctype_AVFrame
	mediaType int
}

// In cause of this:
// > AVFrame is typically allocated once and then reused multiple times to hold
// > different data (e.g. a single AVFrame to hold frames received from a
// > decoder).
// this stuff with map of singletons is used.
//
// @todo find for a better way.
// @todo use something more DRY instead of switch
//
var frames map[int]*Frame = make(map[int]*Frame, 0)

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

func (this *Frame) Encode(cc *CodecCtx) (*Packet, error) {
	var gotOutput int

	p := NewPacket()

	switch this.mediaType {
	case CODEC_TYPE_AUDIO:
		ret := C.avcodec_encode_audio2(cc.avCodecCtx, &p.avPacket, this.avFrame, (*C.int)(unsafe.Pointer(&gotOutput)))
		if ret < 0 {
			return nil, errors.New(fmt.Sprintf("Unable to decode video packet, averror: %s", AvError(int(ret))))
		}

		break

	case CODEC_TYPE_VIDEO:
		ret := C.avcodec_encode_video2(cc.avCodecCtx, &p.avPacket, this.avFrame, (*C.int)(unsafe.Pointer(&gotOutput)))
		if ret < 0 {
			return nil, errors.New(fmt.Sprintf("Unable to decode video packet, averror: %s", AvError(int(ret))))
		}

		break

	default:
		return nil, errors.New(fmt.Sprintf("Unknown codec type: %v", this.mediaType))
	}

	return p, nil
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
