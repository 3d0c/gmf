package gmf

/*

#cgo pkg-config: libavcodec libavutil

#include "libavutil/avutil.h"
#include "libavutil/error.h"
#include "libavutil/mathematics.h"
#include "libavutil/rational.h"

*/
import "C"

import (
	"bytes"
	"errors"
	"unsafe"
)

type AVRational _Ctype_AVRational

type AVR struct {
	Num int
	Den int
}

var (
	AV_TIME_BASE   int        = C.AV_TIME_BASE
	AV_TIME_BASE_Q AVRational = AVRational{1, C.int(AV_TIME_BASE)}
)

func AvError(averr int) error {
	errlen := 1024
	b := make([]byte, errlen)

	C.av_strerror(C.int(averr), (*C.char)(unsafe.Pointer(&b[0])), C.size_t(errlen))

	return errors.New(string(b[:bytes.Index(b, []byte{0})]))
}

func RescaleQ(a int, encBase AVRational, stBase AVRational) int {
	return int(C.av_rescale_q(C.int64_t(a), _Ctype_AVRational(encBase), _Ctype_AVRational(stBase)))
}

func assert(i interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}

	return i
}

// Synthetic video generator. It produces 25 iteratable frames.
// Used for tests.
func GenSyntVideo(w, h int, fmt int32) chan *Frame {
	yield := make(chan *Frame)

	frame := NewFrame().SetWidth(w).SetHeight(h).SetFormat(fmt)

	if err := frame.ImgAlloc(); err != nil {
		return nil
	}

	go func() {
		for i := 0; i < 25; i++ {
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					frame.SetData(0, y*frame.LineSize(0)+x, x+y+i*3)
				}
			}

			// Cb and Cr
			for y := 0; y < h/2; y++ {
				for x := 0; x < w/2; x++ {
					frame.SetData(1, y*frame.LineSize(1)+x, 128+y+i*2)
					frame.SetData(2, y*frame.LineSize(2)+x, 64+x+i*5)
				}
			}

			yield <- frame
		}

		close(yield)
	}()

	return yield
}
