package gmf

/*

#cgo pkg-config: libavcodec

#include "libavcodec/avcodec.h"

*/
import "C"

import (
	"log"
)

type CodecDescriptor struct {
	avDesc    *_Ctype_AVCodecDescriptor
	IsEncoder bool
}

var Codecs []*CodecDescriptor

func InitDesc() {
	var desc *_Ctype_AVCodecDescriptor = nil
	var c *_Ctype_AVCodec

	if Codecs != nil {
		log.Println("Wrong method call. Map 'Codecs' is already initialized. Ignoring...")
		return
	}

	Codecs = make([]*CodecDescriptor, 0)

	for {
		if c = C.av_codec_next(c); c == nil {
			break
		}

		if desc = C.avcodec_descriptor_get(c.id); desc == nil {
			log.Printf("Unable to get descriptor for codec id: %d\n", int(c.id))
		}

		result := &CodecDescriptor{avDesc: desc, IsEncoder: false}

		if C.av_codec_is_encoder(c) > 0 {
			result.IsEncoder = true
		}

		Codecs = append(Codecs, result)
	}

}

func (this *CodecDescriptor) Id() int {
	return int(this.avDesc.id)
}

func (this *CodecDescriptor) Type() int {
	return int(this.avDesc._type)
}

func (this *CodecDescriptor) Name() string {
	return C.GoString(this.avDesc.name)
}

func (this *CodecDescriptor) LongName() string {
	return C.GoString(this.avDesc.name)
}

func (this *CodecDescriptor) Props() int {
	return int(this.avDesc.props)
}
