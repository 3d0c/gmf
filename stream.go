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
}

func (this *Stream) CodecCtx() *CodecCtx {
	if this.cc != nil {
		return this.cc
	}

	c, err := NewDecoder(int(this.avStream.codec.codec_id))
	if err != nil {
		panic(fmt.Sprintf("unable to initialize codec for stream '%d', error:", this.Index(), err))
	}

	this.cc = &CodecCtx{
		codec:      c,
		avCodecCtx: this.avStream.codec,
	}

	return this.cc
}

func (this *Stream) GetCodecCtx() *CodecCtx {
	panic("[deprecated] deprecated method call")
	return nil
}

func (this *Stream) SetCodecCtx(cc *CodecCtx) {
	if cc == nil {
		// don't sure about it.
		panic("Codec context is not initialized.")
	}

	this.avStream.codec = cc.avCodecCtx

	if this.cc != nil {
		this.cc.avCodecCtx = cc.avCodecCtx
	}
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
