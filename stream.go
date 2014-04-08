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

func (this *Stream) GetCodecCtx() *CodecCtx {
	if this.cc != nil {
		return this.cc
	}

	c, err := NewDecoder(int(this.avStream.codec.codec_id))
	if err != nil {
		panic(fmt.Sprintf("Can't init codec for stream '%d', error:", this.Index(), err))
	}

	this.cc = &CodecCtx{
		codec:      c,
		avCodecCtx: this.avStream.codec,
	}

	if err := this.cc.Open(nil); err != nil {
		panic(fmt.Sprintf("Can't open code for stream '%d', error: %v", this.Index(), err))
	}

	return this.cc
}

func (this *Stream) SetCodecCtx(cc *CodecCtx) error {
	// check here, is codec opened.
	this.avStream.codec = cc.avCodecCtx
	return nil
}

func (this *Stream) RescaleTimestamp() int {
	return int(C.av_rescale_q(1, this.avStream.codec.time_base, this.avStream.time_base))
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
