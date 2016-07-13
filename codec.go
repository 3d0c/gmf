package gmf

/*

#cgo pkg-config: libavcodec

#include <stdlib.h>
#include "libavcodec/avcodec.h"
#include "libavutil/pixfmt.h"

*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

var (
	AVMEDIA_TYPE_AUDIO int32 = C.AVMEDIA_TYPE_AUDIO
	AVMEDIA_TYPE_VIDEO int32 = C.AVMEDIA_TYPE_VIDEO

	AV_PIX_FMT_BGR24    int32 = C.AV_PIX_FMT_BGR24
	AV_PIX_FMT_GRAY8    int32 = C.AV_PIX_FMT_GRAY8
	AV_PIX_FMT_RGB24    int32 = C.AV_PIX_FMT_RGB24
	AV_PIX_FMT_YUV410P  int32 = C.AV_PIX_FMT_YUV410P
	AV_PIX_FMT_YUV411P  int32 = C.AV_PIX_FMT_YUV411P
	AV_PIX_FMT_YUV420P  int32 = C.AV_PIX_FMT_YUV420P
	AV_PIX_FMT_YUV422P  int32 = C.AV_PIX_FMT_YUV422P
	AV_PIX_FMT_YUV444P  int32 = C.AV_PIX_FMT_YUV444P
	AV_PIX_FMT_YUVJ420P int32 = C.AV_PIX_FMT_YUVJ420P
	AV_PIX_FMT_YUYV422  int32 = C.AV_PIX_FMT_YUYV422
	AV_PIX_FMT_NONE     int32 = C.AV_PIX_FMT_NONE

	FF_PROFILE_MPEG4_SIMPLE                    int = C.FF_PROFILE_MPEG4_SIMPLE
	FF_PROFILE_MPEG4_SIMPLE_SCALABLE           int = C.FF_PROFILE_MPEG4_SIMPLE_SCALABLE
	FF_PROFILE_MPEG4_CORE                      int = C.FF_PROFILE_MPEG4_CORE
	FF_PROFILE_MPEG4_MAIN                      int = C.FF_PROFILE_MPEG4_MAIN
	FF_PROFILE_MPEG4_N_BIT                     int = C.FF_PROFILE_MPEG4_N_BIT
	FF_PROFILE_MPEG4_SCALABLE_TEXTURE          int = C.FF_PROFILE_MPEG4_SCALABLE_TEXTURE
	FF_PROFILE_MPEG4_SIMPLE_FACE_ANIMATION     int = C.FF_PROFILE_MPEG4_SIMPLE_FACE_ANIMATION
	FF_PROFILE_MPEG4_BASIC_ANIMATED_TEXTURE    int = C.FF_PROFILE_MPEG4_BASIC_ANIMATED_TEXTURE
	FF_PROFILE_MPEG4_HYBRID                    int = C.FF_PROFILE_MPEG4_HYBRID
	FF_PROFILE_MPEG4_ADVANCED_REAL_TIME        int = C.FF_PROFILE_MPEG4_ADVANCED_REAL_TIME
	FF_PROFILE_MPEG4_CORE_SCALABLE             int = C.FF_PROFILE_MPEG4_CORE_SCALABLE
	FF_PROFILE_MPEG4_ADVANCED_CODING           int = C.FF_PROFILE_MPEG4_ADVANCED_CODING
	FF_PROFILE_MPEG4_ADVANCED_CORE             int = C.FF_PROFILE_MPEG4_ADVANCED_CORE
	FF_PROFILE_MPEG4_ADVANCED_SCALABLE_TEXTURE int = C.FF_PROFILE_MPEG4_ADVANCED_SCALABLE_TEXTURE
	FF_PROFILE_MPEG4_SIMPLE_STUDIO             int = C.FF_PROFILE_MPEG4_SIMPLE_STUDIO
	FF_PROFILE_MPEG4_ADVANCED_SIMPLE           int = C.FF_PROFILE_MPEG4_ADVANCED_SIMPLE

	AV_NOPTS_VALUE int64 = C.AV_NOPTS_VALUE
)

func init() {
	C.avcodec_register_all()
	InitDesc()
}

type Codec struct {
	avCodec *C.struct_AVCodec
	CgoMemoryManage
}

func FindDecoder(i interface{}) (*Codec, error) {
	var avc *C.struct_AVCodec

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

func FindEncoder(i interface{}) (*Codec, error) {
	var avc *C.struct_AVCodec

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

func (this *Codec) Free() {
	//nothing to do
}
func (this *Codec) Id() int {
	return int(this.avCodec.id)
}

func (this *Codec) Name() string {
	return C.GoString(this.avCodec.name)
}

func (this *Codec) LongName() string {
	return C.GoString(this.avCodec.long_name)
}

func (this *Codec) Type() int {
	// > ...field names that are keywords in Go can be
	// > accessed by prefixing them with an underscore
	return int(this.avCodec._type)
}

func (this *Codec) IsExperimental() bool {
	return bool((this.avCodec.capabilities & C.CODEC_CAP_EXPERIMENTAL) != 0)
}
