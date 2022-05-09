package gmf

/*

#cgo pkg-config: libavcodec

#include "libavcodec/avcodec.h"

*/
import "C"

type CodecParameters struct {
	avCodecParameters *C.struct_AVCodecParameters
}

func NewCodecParameters() *CodecParameters {
	return &CodecParameters{
		avCodecParameters: C.avcodec_parameters_alloc(),
	}
}

func (cp *CodecParameters) Free() {
	C.avcodec_parameters_free(&cp.avCodecParameters)
}

func (cp *CodecParameters) CodecType() int {
	return int(cp.avCodecParameters.codec_type)
}

func (cp *CodecParameters) CodecId() int {
	return int(cp.avCodecParameters.codec_id)
}

func (cp *CodecParameters) BitRate() int64 {
	return int64(cp.avCodecParameters.bit_rate)
}

func (cp *CodecParameters) Width() int {
	return int(cp.avCodecParameters.width)
}

// Format
// video: the pixel format, the value corresponds to enum AVPixelFormat.
// audio: the sample format, the value corresponds to enum AVSampleFormat.
func (cp *CodecParameters) Format() int32 {
	return int32(cp.avCodecParameters.format)
}

func (cp *CodecParameters) Height() int {
	return int(cp.avCodecParameters.height)
}

func (cp *CodecParameters) FromContext(cc *CodecCtx) error {
	ret := int(C.avcodec_parameters_from_context(cp.avCodecParameters, cc.avCodecCtx))
	if ret < 0 {
		return AvError(ret)
	}

	return nil
}

func (cp *CodecParameters) ToContext(cc *CodecCtx) error {
	ret := int(C.avcodec_parameters_to_context(cc.avCodecCtx, cp.avCodecParameters))
	if ret < 0 {
		return AvError(ret)
	}

	return nil
}
