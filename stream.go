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
	Pts      int64
	CgoMemoryManage
}

func (s *Stream) Free() {
	// nothing to do
}

func (s *Stream) DumpContexCodec(codec *CodecCtx) {
	ret := C.avcodec_copy_context(s.avStream.codec, codec.avCodecCtx)
	if ret < 0 {
		panic("Failed to copy context from input to output stream codec context\n")
	}
}

func (s *Stream) SetCodecFlags() {
	s.avStream.codec.flags |= C.AV_CODEC_FLAG_GLOBAL_HEADER
}

func (s *Stream) CodecCtx() *CodecCtx {
	if s.IsCodecCtxSet() {
		return s.cc
	}

	// @todo make explicit decoder/encoder definition
	// If the codec context wasn't set, it means that it's called from InputCtx
	// and it should be decoder.
	c, err := FindDecoder(int(s.avStream.codec.codec_id))
	if err != nil {
		panic(fmt.Errorf("error initializing codec for stream '%d' - %s", s.Index(), err))
	}

	s.cc = &CodecCtx{
		codec:      c,
		avCodecCtx: s.avStream.codec,
	}

	s.cc.Open(nil)

	return s.cc
}

func (s *Stream) SetCodecCtx(cc *CodecCtx) {
	if cc == nil {
		// don't sure that it should panic...
		panic("Codec context is not initialized.")
	}

	Retain(cc) //just Retain .not need Release,it can free memory by C.avformat_free_context() @ format.go Free().
	s.avStream.codec = cc.avCodecCtx

	if s.cc != nil {
		s.cc.avCodecCtx = cc.avCodecCtx
	}
}

func (s *Stream) IsCodecCtxSet() bool {
	return (s.cc != nil)
}

func (s *Stream) Index() int {
	return int(s.avStream.index)
}

func (s *Stream) Id() int {
	return int(s.avStream.id)
}

func (s *Stream) NbFrames() int {
	if int(s.avStream.nb_frames) == 0 {
		return 1
	}

	return int(s.avStream.nb_frames)
}

func (s *Stream) TimeBase() AVRational {
	return AVRational(s.avStream.time_base)
}

func (s *Stream) Type() int32 {
	return s.CodecCtx().Type()
}

func (s *Stream) IsAudio() bool {
	return (s.Type() == AVMEDIA_TYPE_AUDIO)
}

func (s *Stream) IsVideo() bool {
	return (s.Type() == AVMEDIA_TYPE_VIDEO)
}

func (s *Stream) Duration() int64 {
	return int64(s.avStream.duration)
}
