package gmf

/*

#cgo pkg-config: libavutil

#include "libavutil/samplefmt.h"

*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

//
// Unfinished.
//

type Sample struct {
	data     **C.uint8_t
	linesize *int
	format   SampleFmt
}

func NewSample(nbSamples, nbChannels int, format SampleFmt) error {
	panic("This stuff is unfinished.")
	this := &Sample{format: format}

	if ret := int(C.av_samples_alloc_array_and_samples(
		&this.data,
		(*_Ctype_int)(unsafe.Pointer(&this.linesize)),
		C.int(nbChannels), C.int(nbSamples), int32(format), 0)); ret < 0 {
		return errors.New(fmt.Sprintf("Unable to allocate array and samples: %v", AvError(ret)))
	}

	return nil
}

func (this *Sample) SampleRealloc(nbSamples, nbChannels int) error {
	if ret := int(C.av_samples_alloc(
		this.data,
		(*_Ctype_int)(unsafe.Pointer(&this.linesize)),
		C.int(nbChannels), C.int(nbSamples), int32(this.format), 0)); ret < 0 {
		return errors.New(fmt.Sprintf("Unable to allocate samples: %v", AvError(ret)))
	}

	return nil
}
