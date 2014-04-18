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
	avStream *_Ctype_AVStream
	cc       *CodecCtx
	Pts      int
}

func (this *Stream) CodecCtx() *CodecCtx {
	if this.IsCodecCtxSet() {
		return this.cc
	}

	// @todo make explicit decoder/encoder definition
	// If the codec context wasn't set, means that it's called from InputCtx
	// and it should be decoder only.
	c, err := NewDecoder(int(this.avStream.codec.codec_id))
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
