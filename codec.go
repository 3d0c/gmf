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

	FF_PROFILE_AAC_MAIN      int = C.FF_PROFILE_AAC_MAIN
	FF_PROFILE_AAC_LOW       int = C.FF_PROFILE_AAC_LOW
	FF_PROFILE_AAC_SSR       int = C.FF_PROFILE_AAC_SSR
	FF_PROFILE_AAC_LTP       int = C.FF_PROFILE_AAC_LTP
	FF_PROFILE_AAC_HE        int = C.FF_PROFILE_AAC_HE
	FF_PROFILE_AAC_HE_V2     int = C.FF_PROFILE_AAC_HE_V2
	FF_PROFILE_AAC_LD        int = C.FF_PROFILE_AAC_LD
	FF_PROFILE_AAC_ELD       int = C.FF_PROFILE_AAC_ELD
	FF_PROFILE_MPEG2_AAC_LOW int = C.FF_PROFILE_MPEG2_AAC_LOW
	FF_PROFILE_MPEG2_AAC_HE  int = C.FF_PROFILE_MPEG2_AAC_HE

	FF_PROFILE_DTS         int = C.FF_PROFILE_DTS
	FF_PROFILE_DTS_ES      int = C.FF_PROFILE_DTS_ES
	FF_PROFILE_DTS_96_24   int = C.FF_PROFILE_DTS_96_24
	FF_PROFILE_DTS_HD_HRA  int = C.FF_PROFILE_DTS_HD_HRA
	FF_PROFILE_DTS_HD_MA   int = C.FF_PROFILE_DTS_HD_MA
	FF_PROFILE_DTS_EXPRESS int = C.FF_PROFILE_DTS_EXPRESS

	FF_PROFILE_MPEG2_422          int = C.FF_PROFILE_MPEG2_422
	FF_PROFILE_MPEG2_HIGH         int = C.FF_PROFILE_MPEG2_HIGH
	FF_PROFILE_MPEG2_SS           int = C.FF_PROFILE_MPEG2_SS
	FF_PROFILE_MPEG2_SNR_SCALABLE int = C.FF_PROFILE_MPEG2_SNR_SCALABLE
	FF_PROFILE_MPEG2_MAIN         int = C.FF_PROFILE_MPEG2_MAIN
	FF_PROFILE_MPEG2_SIMPLE       int = C.FF_PROFILE_MPEG2_SIMPLE

	FF_PROFILE_H264_BASELINE            int = C.FF_PROFILE_H264_BASELINE
	FF_PROFILE_H264_MAIN                int = C.FF_PROFILE_H264_MAIN
	FF_PROFILE_H264_EXTENDED            int = C.FF_PROFILE_H264_EXTENDED
	FF_PROFILE_H264_HIGH                int = C.FF_PROFILE_H264_HIGH
	FF_PROFILE_H264_HIGH_10             int = C.FF_PROFILE_H264_HIGH_10
	FF_PROFILE_H264_HIGH_422            int = C.FF_PROFILE_H264_HIGH_422
	FF_PROFILE_H264_HIGH_444            int = C.FF_PROFILE_H264_HIGH_444
	FF_PROFILE_H264_HIGH_444_PREDICTIVE int = C.FF_PROFILE_H264_HIGH_444_PREDICTIVE
	FF_PROFILE_H264_CAVLC_444           int = C.FF_PROFILE_H264_CAVLC_444

	FF_PROFILE_VC1_SIMPLE   int = C.FF_PROFILE_VC1_SIMPLE
	FF_PROFILE_VC1_MAIN     int = C.FF_PROFILE_VC1_MAIN
	FF_PROFILE_VC1_COMPLEX  int = C.FF_PROFILE_VC1_COMPLEX
	FF_PROFILE_VC1_ADVANCED int = C.FF_PROFILE_VC1_ADVANCED

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

	FF_PROFILE_JPEG2000_CSTREAM_RESTRICTION_0  int = C.FF_PROFILE_JPEG2000_CSTREAM_RESTRICTION_0
	FF_PROFILE_JPEG2000_CSTREAM_RESTRICTION_1  int = C.FF_PROFILE_JPEG2000_CSTREAM_RESTRICTION_1
	FF_PROFILE_JPEG2000_CSTREAM_NO_RESTRICTION int = C.FF_PROFILE_JPEG2000_CSTREAM_NO_RESTRICTION
	FF_PROFILE_JPEG2000_DCINEMA_2K             int = C.FF_PROFILE_JPEG2000_DCINEMA_2K
	FF_PROFILE_JPEG2000_DCINEMA_4K             int = C.FF_PROFILE_JPEG2000_DCINEMA_4K

	FF_PROFILE_VP9_0 int = C.FF_PROFILE_VP9_0
	FF_PROFILE_VP9_1 int = C.FF_PROFILE_VP9_1
	FF_PROFILE_VP9_2 int = C.FF_PROFILE_VP9_2
	FF_PROFILE_VP9_3 int = C.FF_PROFILE_VP9_3

	FF_PROFILE_HEVC_MAIN               int = C.FF_PROFILE_HEVC_MAIN
	FF_PROFILE_HEVC_MAIN_10            int = C.FF_PROFILE_HEVC_MAIN_10
	FF_PROFILE_HEVC_MAIN_STILL_PICTURE int = C.FF_PROFILE_HEVC_MAIN_STILL_PICTURE
	FF_PROFILE_HEVC_REXT               int = C.FF_PROFILE_HEVC_REXT
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
