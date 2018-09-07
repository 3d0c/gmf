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

func (cp *CodecParameters) GetCodecType() int {
	return int(cp.avCodecParameters.codec_type)
}

func (cp *CodecParameters) GetCodecId() int {
	return int(cp.avCodecParameters.codec_id)
}

func (cp *CodecParameters) GetBitRate() int64 {
	return int64(cp.avCodecParameters.bit_rate)
}

func (cp *CodecParameters) GetWidth() int {
	return int(cp.avCodecParameters.width)
}

func (cp *CodecParameters) GetHeight() int {
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
