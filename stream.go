package gmf

/*

#cgo pkg-config: libavformat

#include "libavformat/avformat.h"

*/
import "C"

import (
	"fmt"
)

type Stream struct {
	avStream *C.struct_AVStream
	cc       *CodecCtx
	Pts      int
	CgoMemoryManage
}

func (this *Stream) Free() {
	// nothing to do
}

func (this *Stream) DumpContexCodec(codec *CodecCtx) {

	ret := C.avcodec_copy_context(this.avStream.codec, codec.avCodecCtx)
	if ret < 0 {
		panic("Failed to copy context from input to output stream codec context\n")
	}
}

func (this *Stream) SetCodecFlags() {
	this.avStream.codec.flags |= C.CODEC_FLAG_GLOBAL_HEADER
}

func (this *Stream) CodecCtx() *CodecCtx {
	if this.IsCodecCtxSet() {
		return this.cc
	}

	// @todo make explicit decoder/encoder definition
	// If the codec context wasn't set, it means that it's called from InputCtx
	// and it should be decoder.
	c, err := FindDecoder(int(this.avStream.codec.codec_id))
	if err != nil {
		panic(fmt.Sprintf("unable to initialize codec for stream '%d', error:", this.Index(), err))
	}

	this.cc = &CodecCtx{
		codec:      c,
		avCodecCtx: this.avStream.codec,
	}

	this.cc.Open(nil)

	return this.cc
}

func (this *Stream) SetCodecCtx(cc *CodecCtx) {
	if cc == nil {
		// don't sure that it should panic...
		panic("Codec context is not initialized.")
	}

	Retain(cc) //just Retain .not need Release,it can free memory by C.avformat_free_context() @ format.go Free().
	this.avStream.codec = cc.avCodecCtx

	if this.cc != nil {
		this.cc.avCodecCtx = cc.avCodecCtx
	}
}

func (this *Stream) IsCodecCtxSet() bool {
	return (this.cc != nil)
}

func (this *Stream) Index() int {
	return int(this.avStream.index)
}

func (this *Stream) Id() int {
	return int(this.avStream.id)
}

func (this *Stream) NbFrames() int {
	return int(this.avStream.nb_frames)
}

func (this *Stream) TimeBase() AVRational {
	return AVRational(this.avStream.time_base)
}

func (this *Stream) Type() int32 {
	return this.CodecCtx().Type()
}

func (this *Stream) IsAudio() bool {
	return (this.Type() == AVMEDIA_TYPE_AUDIO)
}

func (this *Stream) IsVideo() bool {
	return (this.Type() == AVMEDIA_TYPE_VIDEO)
}

func (this *Stream) Duration() int64 {
	return int64(this.avStream.duration)
}
