// +build go1.12

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

type SampleFormat int

type Sample struct {
	data     **C.uint8_t
	linesize *int
	format   SampleFormat
	CgoMemoryManage
}

func NewSample(nbSamples, nbChannels int, format SampleFormat) error {
	panic("This stuff is unfinished.")
	sample := &Sample{format: format}

	if ret := int(C.av_samples_alloc_array_and_samples(
		&sample.data,
		(*C.int)(unsafe.Pointer(&sample.linesize)),
		C.int(nbChannels), C.int(nbSamples), int32(format), 0)); ret < 0 {
		return errors.New(fmt.Sprintf("Unable to allocate array and samples: %v", AvError(ret)))
	}

	return nil
}

func (s *Sample) SampleRealloc(nbSamples, nbChannels int) error {
	if ret := int(C.av_samples_alloc(
		s.data,
		(*C.int)(unsafe.Pointer(&s.linesize)),
		C.int(nbChannels), C.int(nbSamples), int32(s.format), 0)); ret < 0 {
		return errors.New(fmt.Sprintf("Unable to allocate samples: %v", AvError(ret)))
	}

	return nil
}
