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
	SwsCtx   *SwsCtx
	SwrCtx   *SwrCtx
	AvFifo   *AVAudioFifo
	DstFrame *Frame
	CgoMemoryManage
}

func (s *Stream) Free() {
	if s.SwsCtx != nil {
		s.SwsCtx.Free()
	}
	if s.DstFrame != nil {
		s.DstFrame.Free()
	}
	if s.SwrCtx != nil {
		s.SwrCtx.Free()
	}
	if s.AvFifo != nil {
		s.AvFifo.Free()
	}
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

	c, err := FindDecoder(int(s.avStream.codec.codec_id))
	if err != nil {
		return nil
		// return fmt.Errorf("error initializing codec for stream '%d' - %s", s.Index(), err)
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
		panic("Codec context is not initialized.")
	}

	Retain(cc) //just Retain .not need Release,it can free memory by C.avformat_free_context() @ format.go Free().
	s.avStream.codec = cc.avCodecCtx

	if s.cc != nil {
		s.cc.avCodecCtx = cc.avCodecCtx
	}
}

func (s *Stream) SetCodecParameters(cp *CodecParameters) error {
	if cp == nil || cp.avCodecParameters == nil {
		return fmt.Errorf("codec parameters are not initialized")
	}

	s.avStream.codecpar = cp.avCodecParameters
	return nil
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

func (s *Stream) SetTimeBase(val AVR) *Stream {
	s.avStream.time_base.num = C.int(val.Num)
	s.avStream.time_base.den = C.int(val.Den)
	return s
}

func (s *Stream) GetRFrameRate() AVRational {
	return AVRational(s.avStream.r_frame_rate)
}

func (s *Stream) SetRFrameRate(val AVR) {
	s.avStream.r_frame_rate.num = C.int(val.Num)
	s.avStream.r_frame_rate.den = C.int(val.Den)
}

func (s *Stream) GetAvgFrameRate() AVRational {
	return AVRational(s.avStream.avg_frame_rate)
}

func (s *Stream) GetStartTime() int64 {
	return int64(s.avStream.start_time)
}

func (s *Stream) GetCodecPar() *CodecParameters {
	cp := NewCodecParameters()
	cp.avCodecParameters = s.avStream.codecpar

	return cp
}

func (s *Stream) CopyCodecPar(cp *CodecParameters) error {
	ret := int(C.avcodec_parameters_copy(s.avStream.codecpar, cp.avCodecParameters))
	if ret < 0 {
		return AvError(ret)
	}

	return nil
}
