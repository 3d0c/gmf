package gmf

/*

#cgo pkg-config: libavformat libavcodec

#include "libavformat/avformat.h"
#include "libavcodec/avcodec.h"

*/
import "C"

import (
	"fmt"
)

type Stream struct {
	avStream *C.struct_AVStream
	SwsCtx   *SwsCtx
	SwrCtx   *SwrCtx
	AvFifo   *AVAudioFifo
	// Deprecated: Using AVStream.codec to pass codec parameters to muxers is deprecated, use AVStream.codecpar instead.
	cc  *CodecCtx
	Pts int64
	CgoMemoryManage
}

func (s *Stream) Free() {
	if s.SwsCtx != nil {
		s.SwsCtx.Free()
	}
	if s.SwrCtx != nil {
		s.SwrCtx.Free()
	}
	if s.AvFifo != nil {
		s.AvFifo.Free()
	}
}

func (s *Stream) DumpContexCodec(codec *CodecCtx) {
	ret := C.avcodec_parameters_from_context(s.avStream.codecpar, codec.avCodecCtx)
	if ret < 0 {
		panic("Failed to copy context from input to output stream codec context\n")
	}
}

func (s *Stream) SetCodecFlags() {
	s.avStream.event_flags |= C.AV_CODEC_FLAG_GLOBAL_HEADER
}

// CodecCtx
// Deprecated: Using AVStream.codec to pass codec parameters to muxers is deprecated, use AVStream.codecpar instead.
func (s *Stream) CodecCtx() *CodecCtx {
	// Supposed that output context is set and opened by user
	if s.IsCodecCtxSet() {
		return s.cc
	}

	// Open input codec context
	c, err := FindDecoder(int(s.avStream.codecpar.codec_id))
	if err != nil {
		return nil
	}

	if s.cc = NewCodecCtx(c); s.cc == nil {
		panic("error allocating codec context")
	}

	ret := int(C.avcodec_parameters_to_context(s.cc.avCodecCtx, s.avStream.codecpar))
	if ret < 0 {
		panic("error copying parameters to codec context")
	}

	if err := s.cc.Open(nil); err != nil {
		panic("error opening codec context")
	}

	s.cc.avCodecCtx.time_base = s.avStream.time_base

	return s.cc
}

// SetCodecCtx
// Deprecated: Using AVStream.codec to pass codec parameters to muxers is deprecated, use AVStream.codecpar instead.
func (s *Stream) SetCodecCtx(cc *CodecCtx) {
	s.cc = cc
}

func (s *Stream) CodecPar() *CodecParameters {
	if s.avStream == nil || s.avStream.codecpar == nil {
		return nil
	}
	return &CodecParameters{
		avCodecParameters: s.avStream.codecpar,
	}
}

func (s *Stream) SetCodecPar(cp *CodecParameters) error {
	if cp == nil || cp.avCodecParameters == nil {
		return fmt.Errorf("codec parameters are not initialized")
	}

	s.avStream.codecpar = cp.avCodecParameters
	return nil
}

// IsCodecCtxSet
// Deprecated: Using AVStream.codec to pass codec parameters to muxers is deprecated, use AVStream.codecpar instead.
func (s *Stream) IsCodecCtxSet() bool {
	return s.cc != nil
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
	return int32(s.CodecPar().CodecType())
}

func (s *Stream) IsAudio() bool {
	return s.Type() == AVMEDIA_TYPE_AUDIO
}

func (s *Stream) IsVideo() bool {
	return s.Type() == AVMEDIA_TYPE_VIDEO
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

func (s *Stream) SetAvgFrameRate(val AVR) {
	s.avStream.avg_frame_rate.num = C.int(val.Num)
	s.avStream.avg_frame_rate.den = C.int(val.Den)
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
