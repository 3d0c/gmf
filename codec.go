package gmf

/*

#cgo pkg-config: libavcodec

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
	CODEC_TYPE_AUDIO   int   = C.AVMEDIA_TYPE_AUDIO
	CODEC_TYPE_VIDEO   int   = C.AVMEDIA_TYPE_VIDEO
	AV_PIX_FMT_YUV420P int32 = C.AV_PIX_FMT_YUV420P
)

func init() {
	C.avcodec_register_all()
}

type Codec struct {
	avCodec *_Ctype_AVCodec
}

func NewDecoder(i interface{}) (*Codec, error) {
	var avc *_Ctype_AVCodec

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

func NewEncoder(i interface{}) (*Codec, error) {
	var avc *_Ctype_AVCodec

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

func (this *Codec) Name() string {
	return C.GoString(this.avCodec.name)
}

func (this *Codec) LongName() string {
	return C.GoString(this.avCodec.long_name)
}

// > ...field names that are keywords in Go can be
// > accessed by prefixing them with an underscore
func (this *Codec) Type() int {
	return int(this.avCodec._type)
}
